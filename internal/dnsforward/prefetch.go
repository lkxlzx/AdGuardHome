package dnsforward

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/miekg/dns"
)

// RefreshTask represents a domain refresh task with priority.
type RefreshTask struct {
	Domain      string
	ExpireTime  time.Time
	Priority    int
	EnqueueTime time.Time
}

// PrefetchMetrics contains statistics about prefetch operations.
type PrefetchMetrics struct {
	CurrentActive   atomic.Int32 // Current active tasks
	SoftLimitHits   atomic.Int64 // Times soft limit was hit
	HardLimitHits   atomic.Int64 // Times hard limit was hit
	UrgentQueueSize atomic.Int32 // Urgent queue length
	NormalQueueSize atomic.Int32 // Normal queue length
	TasksUpgraded   atomic.Int64 // Tasks upgraded to urgent
	TasksDropped    atomic.Int64 // Tasks dropped
	TasksCompleted  atomic.Int64 // Completed tasks
	TasksFailed     atomic.Int64 // Failed tasks
}

// domainShard represents a shard of the domain tracking maps.
// Using sharded maps reduces lock contention in high-concurrency scenarios.
type domainShard struct {
	mu         sync.RWMutex
	domains    map[string]time.Time
	hits       map[string]int
	lastAccess map[string]time.Time
	// hitTimestamps stores the timestamps of each hit for time window calculation
	hitTimestamps map[string][]time.Time
}

// PrefetchManager handles active cache warming for hot domains.
type PrefetchManager struct {
	logger *slog.Logger
	server *Server

	// shards is an array of domain shards for reduced lock contention.
	// Using 16 shards provides good balance between memory and concurrency.
	shards [16]*domainShard

	// threshold is the number of hits required to consider a domain "hot".
	threshold int

	// timeWindow is the time window for counting hits.
	// Only hits within this window are counted towards the threshold.
	timeWindow time.Duration

	// maxEntries is the maximum number of entries to track.
	// When exceeded, cleanup will be triggered.
	maxEntries int

	// cleanupInterval is the interval between cleanup operations.
	cleanupInterval time.Duration

	// Dynamic concurrency control
	softLimit        int // Soft limit for concurrent operations
	hardLimit        int // Hard limit for concurrent operations
	urgentThreshold  int // Priority threshold for urgent tasks

	// Task queues
	urgentQueue chan *RefreshTask // High priority queue
	normalQueue chan *RefreshTask // Normal priority queue

	// Metrics
	metrics *PrefetchMetrics

	// stopCh is used to signal the worker to stop.
	stopCh chan struct{}

	// Worker control
	workerWg sync.WaitGroup
}

// NewPrefetchManager creates a new PrefetchManager with configuration from server.
func NewPrefetchManager(s *Server) *PrefetchManager {
	conf := s.conf

	// Apply configuration with defaults
	threshold := conf.PrefetchThreshold
	if threshold <= 0 || threshold > 100 {
		threshold = 5 // Default: 5 hits
	}

	timeWindow := time.Duration(conf.PrefetchTimeWindow)
	if timeWindow <= 0 {
		timeWindow = 1 * time.Hour // Default: 1 hour
	}

	maxEntries := conf.PrefetchMaxEntries
	if maxEntries < 1000 || maxEntries > 100000 {
		maxEntries = 10000 // Default: 10,000 entries
	}

	cleanupInterval := time.Duration(conf.PrefetchCleanupInterval)
	if cleanupInterval <= 0 {
		cleanupInterval = 1 * time.Hour // Default: 1 hour
	}

	// Dynamic concurrency configuration
	softLimit := conf.PrefetchSoftLimit
	if softLimit <= 0 {
		// Fallback to old config for backward compatibility
		softLimit = conf.PrefetchMaxConcurrentRefresh
	}
	if softLimit < 10 || softLimit > 500 {
		softLimit = 50 // Default: 50
	}

	hardLimit := conf.PrefetchHardLimit
	if hardLimit <= 0 {
		hardLimit = softLimit * 3 // Default: 3x soft limit
	}
	if hardLimit < 50 || hardLimit > 1000 {
		hardLimit = 150 // Default: 150
	}
	if hardLimit < softLimit {
		hardLimit = softLimit * 2 // Ensure hard limit > soft limit
	}

	urgentQueueSize := conf.PrefetchUrgentQueueSize
	if urgentQueueSize <= 0 {
		urgentQueueSize = 500 // Default: 500
	}

	normalQueueSize := conf.PrefetchNormalQueueSize
	if normalQueueSize <= 0 {
		normalQueueSize = 2000 // Default: 2000
	}

	urgentThreshold := conf.PrefetchUrgentThreshold
	if urgentThreshold <= 0 || urgentThreshold > 90 {
		urgentThreshold = 70 // Default: 70
	}

	pm := &PrefetchManager{
		logger:          s.logger.With(slogutil.KeyPrefix, "prefetch"),
		server:          s,
		threshold:       threshold,
		timeWindow:      timeWindow,
		maxEntries:      maxEntries,
		cleanupInterval: cleanupInterval,
		softLimit:       softLimit,
		hardLimit:       hardLimit,
		urgentThreshold: urgentThreshold,
		urgentQueue:     make(chan *RefreshTask, urgentQueueSize),
		normalQueue:     make(chan *RefreshTask, normalQueueSize),
		metrics:         &PrefetchMetrics{},
		stopCh:          make(chan struct{}),
	}

	// Initialize shards
	for i := range pm.shards {
		pm.shards[i] = &domainShard{
			domains:       make(map[string]time.Time),
			hits:          make(map[string]int),
			lastAccess:    make(map[string]time.Time),
			hitTimestamps: make(map[string][]time.Time),
		}
	}

	s.logger.Info("prefetch manager initialized",
		"threshold", threshold,
		"time_window", timeWindow,
		"max_entries", maxEntries,
		"cleanup_interval", cleanupInterval,
		"soft_limit", softLimit,
		"hard_limit", hardLimit,
		"urgent_threshold", urgentThreshold,
		"urgent_queue_size", urgentQueueSize,
		"normal_queue_size", normalQueueSize)

	return pm
}

// Start starts the background workers.
func (pm *PrefetchManager) Start() {
	// Start multiple workers for better concurrency
	numWorkers := pm.softLimit / 10
	if numWorkers < 2 {
		numWorkers = 2
	}
	if numWorkers > 10 {
		numWorkers = 10
	}

	pm.logger.Info("starting prefetch workers", "count", numWorkers)

	for i := 0; i < numWorkers; i++ {
		pm.workerWg.Add(1)
		go pm.worker(i)
	}

	// Start cleanup worker
	pm.workerWg.Add(1)
	go pm.cleanupWorker()

	// Start metrics logger
	pm.workerWg.Add(1)
	go pm.metricsLogger()
}

// Stop stops all background workers.
func (pm *PrefetchManager) Stop() {
	close(pm.stopCh)
	pm.workerWg.Wait()
	pm.logger.Info("prefetch manager stopped")
}

// getShard returns the shard for a given domain using a simple hash function.
func (pm *PrefetchManager) getShard(domain string) *domainShard {
	// Simple hash function: sum of bytes modulo number of shards
	var hash uint32
	for i := 0; i < len(domain); i++ {
		hash = hash*31 + uint32(domain[i])
	}
	return pm.shards[hash%uint32(len(pm.shards))]
}

// getTotalEntries returns the total number of entries across all shards.
func (pm *PrefetchManager) getTotalEntries() int {
	total := 0
	for _, shard := range pm.shards {
		shard.mu.RLock()
		total += len(shard.hits)
		shard.mu.RUnlock()
	}
	return total
}

// Record updates the hit count and expiration time for a domain.
func (pm *PrefetchManager) Record(domain string, ttl uint32) {
	if domain == "" || ttl == 0 {
		return
	}

	domain = dns.Fqdn(domain)
	now := time.Now()

	// Get the shard for this domain
	shard := pm.getShard(domain)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Add current timestamp to hit history
	timestamps := shard.hitTimestamps[domain]
	timestamps = append(timestamps, now)

	// Remove timestamps outside the time window
	cutoff := now.Add(-pm.timeWindow)
	validTimestamps := make([]time.Time, 0, len(timestamps))
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	shard.hitTimestamps[domain] = validTimestamps

	// Update hit count based on valid timestamps within time window
	hitCount := len(validTimestamps)
	shard.hits[domain] = hitCount
	shard.lastAccess[domain] = now

	// Only track if it exceeds the threshold
	if hitCount >= pm.threshold {
		// Calculate expiration time
		shard.domains[domain] = now.Add(time.Duration(ttl) * time.Second)
	} else {
		// Remove from hot domains if it falls below threshold
		delete(shard.domains, domain)
	}

	// Check if we need cleanup (check periodically, not every time)
	// Only check on shard 0 to avoid multiple cleanup triggers
	if shard == pm.shards[0] && len(shard.hits)%1000 == 0 {
		totalEntries := pm.getTotalEntries()
		if totalEntries > pm.maxEntries {
			pm.logger.Info("prefetch cache exceeded max entries, triggering cleanup",
				"current", totalEntries,
				"max", pm.maxEntries)
			// Trigger async cleanup to avoid blocking
			go pm.cleanup()
		}
	}
}

// calculatePriority calculates the priority of a refresh task based on time until expiration.
// Returns 0-100, where higher values indicate higher priority.
func (pm *PrefetchManager) calculatePriority(expireTime time.Time, now time.Time) int {
	timeUntilExpire := expireTime.Sub(now)

	if timeUntilExpire <= 0 {
		return 100 // Already expired, highest priority
	}

	if timeUntilExpire <= 5*time.Second {
		return 90 // Expiring in 5 seconds, very high priority
	}

	if timeUntilExpire <= 30*time.Second {
		return 70 // Expiring in 30 seconds, high priority
	}

	if timeUntilExpire <= 2*time.Minute {
		return 50 // Expiring in 2 minutes, medium priority
	}

	return 30 // Normal priority
}

// scheduleTask adds a task to the appropriate queue based on priority.
func (pm *PrefetchManager) scheduleTask(task *RefreshTask) {
	task.Priority = pm.calculatePriority(task.ExpireTime, time.Now())
	task.EnqueueTime = time.Now()

	if task.Priority >= pm.urgentThreshold {
		select {
		case pm.urgentQueue <- task:
			pm.metrics.UrgentQueueSize.Add(1)
		default:
			pm.logger.Warn("urgent queue full, dropping task", "domain", task.Domain)
			pm.metrics.TasksDropped.Add(1)
		}
	} else {
		select {
		case pm.normalQueue <- task:
			pm.metrics.NormalQueueSize.Add(1)
		default:
			pm.logger.Warn("normal queue full, dropping task", "domain", task.Domain)
			pm.metrics.TasksDropped.Add(1)
		}
	}
}

// worker processes tasks from both queues with dynamic concurrency control.
func (pm *PrefetchManager) worker(id int) {
	defer pm.workerWg.Done()

	pm.logger.Debug("worker started", "id", id)

	for {
		select {
		case <-pm.stopCh:
			pm.logger.Debug("worker stopped", "id", id)
			return

		case task := <-pm.urgentQueue:
			pm.metrics.UrgentQueueSize.Add(-1)
			pm.executeTask(task, true)

		case task := <-pm.normalQueue:
			pm.metrics.NormalQueueSize.Add(-1)

			// Check soft limit for normal tasks
			current := pm.metrics.CurrentActive.Load()
			if current >= int32(pm.softLimit) {
				pm.metrics.SoftLimitHits.Add(1)

				// Re-evaluate priority
				newPriority := pm.calculatePriority(task.ExpireTime, time.Now())
				if newPriority >= pm.urgentThreshold {
					// Upgrade to urgent
					pm.metrics.TasksUpgraded.Add(1)
					pm.logger.Debug("task upgraded to urgent",
						"domain", task.Domain,
						"old_priority", task.Priority,
						"new_priority", newPriority)

					task.Priority = newPriority
					select {
					case pm.urgentQueue <- task:
						pm.metrics.UrgentQueueSize.Add(1)
					default:
						pm.logger.Warn("urgent queue full after upgrade", "domain", task.Domain)
						pm.metrics.TasksDropped.Add(1)
					}
					continue
				}

				// Wait a bit and re-queue
				time.Sleep(100 * time.Millisecond)
				select {
				case pm.normalQueue <- task:
					pm.metrics.NormalQueueSize.Add(1)
				default:
					pm.logger.Warn("normal queue full on re-queue", "domain", task.Domain)
					pm.metrics.TasksDropped.Add(1)
				}
				continue
			}

			pm.executeTask(task, false)
		}
	}
}

// executeTask executes a refresh task with concurrency control.
func (pm *PrefetchManager) executeTask(task *RefreshTask, isUrgent bool) {
	current := pm.metrics.CurrentActive.Add(1)
	defer pm.metrics.CurrentActive.Add(-1)

	// Check hard limit
	if current > int32(pm.hardLimit) {
		pm.metrics.HardLimitHits.Add(1)
		pm.logger.Warn("exceeded hard limit",
			"current", current,
			"limit", pm.hardLimit,
			"domain", task.Domain,
			"urgent", isUrgent)

		// Drop non-urgent tasks that exceed hard limit
		if !isUrgent {
			pm.metrics.TasksDropped.Add(1)
			return
		}
	}

	// Execute refresh in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				pm.logger.Error("task panic", "domain", task.Domain, "error", r)
				pm.metrics.TasksFailed.Add(1)
			}
		}()

		// Record wait time
		waitTime := time.Since(task.EnqueueTime)
		if waitTime > 5*time.Second {
			pm.logger.Warn("long wait time",
				"domain", task.Domain,
				"wait", waitTime,
				"priority", task.Priority)
		}

		// Perform refresh
		err := pm.refresh(task.Domain)
		if err != nil {
			pm.logger.Warn("refresh failed", "domain", task.Domain, "err", err)
			pm.metrics.TasksFailed.Add(1)
		} else {
			pm.metrics.TasksCompleted.Add(1)
			pm.logger.Debug("refresh completed",
				"domain", task.Domain,
				"wait", waitTime,
				"priority", task.Priority)
		}
	}()
}

// cleanupWorker periodically scans for domains that need refreshing.
func (pm *PrefetchManager) cleanupWorker() {
	defer pm.workerWg.Done()

	refreshTicker := time.NewTicker(10 * time.Second)
	cleanupTicker := time.NewTicker(pm.cleanupInterval)
	defer refreshTicker.Stop()
	defer cleanupTicker.Stop()

	pm.logger.Debug("cleanup worker started")

	for {
		select {
		case <-pm.stopCh:
			pm.logger.Debug("cleanup worker stopped")
			return

		case <-refreshTicker.C:
			pm.checkAndRefresh()

		case <-cleanupTicker.C:
			pm.cleanup()
		}
	}
}

// checkAndRefresh scans for expired domains and schedules refresh tasks.
func (pm *PrefetchManager) checkAndRefresh() {
	now := time.Now()
	var tasks []*RefreshTask

	// Scan all shards
	for _, shard := range pm.shards {
		shard.mu.Lock()
		for domain, expiry := range shard.domains {
			// If expired or about to expire (within 5 seconds)
			if now.After(expiry) || now.Add(5*time.Second).After(expiry) {
				tasks = append(tasks, &RefreshTask{
					Domain:     domain,
					ExpireTime: expiry,
				})
				// Remove from domains map to prevent duplicate scheduling
				delete(shard.domains, domain)
			}
		}
		shard.mu.Unlock()
	}

	if len(tasks) == 0 {
		return
	}

	pm.logger.Debug("scheduling refresh tasks", "count", len(tasks))

	// Schedule all tasks
	for _, task := range tasks {
		pm.scheduleTask(task)
	}
}

// cleanup removes stale entries from hits, domains, and lastAccess maps
// to prevent memory leaks. This is called periodically and when maxEntries is exceeded.
func (pm *PrefetchManager) cleanup() {
	now := time.Now()
	cleanupThreshold := 24 * time.Hour // Remove entries not accessed in 24 hours
	lowHitThreshold := pm.threshold    // Remove entries below threshold

	var totalRemoved int
	var totalHits, totalDomains int

	// Clean up each shard
	for _, shard := range pm.shards {
		shard.mu.Lock()

		var removed int

		// Clean up entries that haven't been accessed recently or have low hit counts
		for domain, lastAccess := range shard.lastAccess {
			shouldRemove := false

			// Remove if not accessed in 24 hours AND not in active domains
			// (keep hot domains even if old)
			if now.Sub(lastAccess) > cleanupThreshold {
				if _, inDomains := shard.domains[domain]; !inDomains {
					shouldRemove = true
				}
			}

			// Remove if hit count is below threshold and not in active domains
			if _, inDomains := shard.domains[domain]; !inDomains {
				if hits, exists := shard.hits[domain]; exists && hits < lowHitThreshold {
					shouldRemove = true
				}
			}

			if shouldRemove {
				delete(shard.hits, domain)
				delete(shard.domains, domain)
				delete(shard.lastAccess, domain)
				delete(shard.hitTimestamps, domain)
				removed++
			}
		}

		totalRemoved += removed
		totalHits += len(shard.hits)
		totalDomains += len(shard.domains)

		shard.mu.Unlock()
	}

	if totalRemoved > 0 {
		pm.logger.Info("prefetch cleanup completed",
			"removed", totalRemoved,
			"remaining_hits", totalHits,
			"remaining_domains", totalDomains)
	}
}

// refresh sends a DNS query to the server itself to trigger a cache update.
func (pm *PrefetchManager) refresh(domain string) error {
	// Create a local DNS client
	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeA)
	m.RecursionDesired = true

	// Using network loopback to simulate a real client and trigger full processing chain.
	port := "53"
	if len(pm.server.conf.UDPListenAddrs) > 0 {
		port = fmt.Sprintf("%d", pm.server.conf.UDPListenAddrs[0].Port)
	}

	// Use 127.0.0.1
	target := net.JoinHostPort("127.0.0.1", port)

	_, _, err := c.Exchange(m, target)
	return err
}

// GetStats returns statistics about the prefetch manager.
// This is useful for monitoring and debugging.
func (pm *PrefetchManager) GetStats() (hits, domains, tracked int) {
	for _, shard := range pm.shards {
		shard.mu.RLock()
		hits += len(shard.hits)
		domains += len(shard.domains)
		tracked += len(shard.lastAccess)
		shard.mu.RUnlock()
	}
	
	return hits, domains, tracked
}

// metricsLogger periodically logs metrics for monitoring.
func (pm *PrefetchManager) metricsLogger() {
	defer pm.workerWg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.logMetrics()
		}
	}
}

// logMetrics logs current metrics.
func (pm *PrefetchManager) logMetrics() {
	hits, domains, tracked := pm.GetStats()

	pm.logger.Info("prefetch metrics",
		"active_tasks", pm.metrics.CurrentActive.Load(),
		"urgent_queue", pm.metrics.UrgentQueueSize.Load(),
		"normal_queue", pm.metrics.NormalQueueSize.Load(),
		"soft_limit_hits", pm.metrics.SoftLimitHits.Load(),
		"hard_limit_hits", pm.metrics.HardLimitHits.Load(),
		"tasks_upgraded", pm.metrics.TasksUpgraded.Load(),
		"tasks_dropped", pm.metrics.TasksDropped.Load(),
		"tasks_completed", pm.metrics.TasksCompleted.Load(),
		"tasks_failed", pm.metrics.TasksFailed.Load(),
		"tracked_hits", hits,
		"hot_domains", domains,
		"tracked_domains", tracked)
}

// GetMetrics returns a snapshot of current metrics.
func (pm *PrefetchManager) GetMetrics() map[string]int64 {
	hits, domains, tracked := pm.GetStats()

	return map[string]int64{
		"current_active":   int64(pm.metrics.CurrentActive.Load()),
		"urgent_queue":     int64(pm.metrics.UrgentQueueSize.Load()),
		"normal_queue":     int64(pm.metrics.NormalQueueSize.Load()),
		"soft_limit_hits":  pm.metrics.SoftLimitHits.Load(),
		"hard_limit_hits":  pm.metrics.HardLimitHits.Load(),
		"tasks_upgraded":   pm.metrics.TasksUpgraded.Load(),
		"tasks_dropped":    pm.metrics.TasksDropped.Load(),
		"tasks_completed":  pm.metrics.TasksCompleted.Load(),
		"tasks_failed":     pm.metrics.TasksFailed.Load(),
		"tracked_hits":     int64(hits),
		"hot_domains":      int64(domains),
		"tracked_domains":  int64(tracked),
	}
}
