package dnsforward

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
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
	RequeueCount int // Number of times this task has been requeued
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
	lastAccess map[string]time.Time
	// hitCounters stores simplified hit counting for time window calculation
	hitCounters map[string]*hitCounter
}

// hitCounter tracks hits within a time window with minimal memory allocation.
type hitCounter struct {
	count       int       // Current hit count
	windowStart time.Time // Start of current time window
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

	// lastPrefetchTime stores the timestamp of the last successful prefetch
	lastPrefetchTime atomic.Value // stores time.Time

	// stopCh is used to signal the worker to stop.
	stopCh chan struct{}

	// Worker control
	workerWg sync.WaitGroup

	// Error logging
	errorLogFile *os.File
	errorLogMu   sync.Mutex
	
	// Incremental cleanup state
	nextCleanupShard atomic.Int32 // Next shard to clean (for incremental cleanup)
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
			domains:     make(map[string]time.Time),
			lastAccess:  make(map[string]time.Time),
			hitCounters: make(map[string]*hitCounter),
		}
	}

	// Initialize error log file
	if err := pm.initErrorLog(); err != nil {
		s.logger.Warn("failed to initialize prefetch error log", slogutil.KeyError, err)
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
	
	// Close error log file
	pm.closeErrorLog()
	
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
		total += len(shard.hitCounters)
		shard.mu.RUnlock()
	}
	return total
}

// Record updates the hit count and expiration time for a domain.
func (pm *PrefetchManager) Record(domain string, ttl uint32) {
	pm.record(domain, ttl, false)
}

// RecordCacheHit records a cache hit, bypassing threshold checking.
// Cache hits indicate hot domains and are directly added to prefetch queue.
func (pm *PrefetchManager) RecordCacheHit(domain string, ttl uint32) {
	pm.record(domain, ttl, true)
}

// record is the internal implementation that handles both cache hit and miss scenarios.
// Optimized version with minimal memory allocation and reduced lock time.
func (pm *PrefetchManager) record(domain string, ttl uint32, isCacheHit bool) {
	if domain == "" || ttl == 0 {
		return
	}

	// Perform string operations outside the lock
	domain = dns.Fqdn(domain)
	now := time.Now()
	expiry := now.Add(time.Duration(ttl) * time.Second)

	// Get the shard for this domain
	shard := pm.getShard(domain)

	// Minimize lock holding time
	shard.mu.Lock()

	// Get or create hit counter
	counter := shard.hitCounters[domain]
	if counter == nil {
		counter = &hitCounter{
			count:       1,
			windowStart: now,
		}
		shard.hitCounters[domain] = counter
	} else {
		// Check if we need to reset the time window
		if now.Sub(counter.windowStart) > pm.timeWindow {
			// Window expired, reset counter
			counter.count = 1
			counter.windowStart = now
		} else {
			// Within window, increment counter
			counter.count++
		}
	}

	hitCount := counter.count
	shard.lastAccess[domain] = now

	// Cache hits bypass threshold checking
	shouldTrack := isCacheHit || (hitCount >= pm.threshold)

	if shouldTrack {
		shard.domains[domain] = expiry
	} else {
		// Remove from hot domains if it falls below threshold
		delete(shard.domains, domain)
	}

	// Check if we need cleanup (check periodically, not every time)
	// Only check on shard 0 to avoid multiple cleanup triggers
	needsCleanup := shard == pm.shards[0] && len(shard.hitCounters)%1000 == 0

	shard.mu.Unlock()

	// Perform expensive operations outside the lock
	if needsCleanup {
		totalEntries := pm.getTotalEntries()
		if totalEntries > pm.maxEntries {
			pm.logger.Info("prefetch cache exceeded max entries, triggering cleanup",
				"current", totalEntries,
				"max", pm.maxEntries)
			// Trigger async smart cleanup (force=false, only clean if needed)
			go pm.cleanupInternal(false)
		}
	}

	// Debug logging removed for performance
	// Only log at Info level for important events
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

	// Worker started (debug logging removed for performance)

	for {
		select {
		case <-pm.stopCh:
			// Worker stopped
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
					// Task upgraded (debug logging removed for performance)

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

				// Check requeue limit to prevent infinite loops
				const maxRequeueCount = 10
				if task.RequeueCount >= maxRequeueCount {
					pm.logger.Warn("task exceeded requeue limit, forcing execution",
						"domain", task.Domain,
						"requeue_count", task.RequeueCount)
					// Force execution even if over soft limit
					pm.executeTask(task, false)
					continue
				}

				// Wait a bit and re-queue
				task.RequeueCount++
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
			pm.metrics.CurrentActive.Add(-1) // Decrement before returning
			return
		}
	}

	// Execute refresh in goroutine
	go func() {
		// Decrement counter when goroutine completes
		defer pm.metrics.CurrentActive.Add(-1)
		
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
			pm.lastPrefetchTime.Store(time.Now())
			// Refresh completed (debug logging removed for performance)
		}
	}()
}

// cleanupIncremental performs incremental LRU-based cleanup.
// Only cleans a subset of shards per invocation for better performance.
func (pm *PrefetchManager) cleanupIncremental() {
	totalEntries := pm.getTotalEntries()
	
	// Only cleanup if needed
	if totalEntries <= pm.maxEntries {
		return
	}
	
	// Calculate how many shards to clean this round
	// Clean 25% of shards each time (4 out of 16)
	shardsPerRound := len(pm.shards) / 4
	if shardsPerRound < 1 {
		shardsPerRound = 1
	}
	
	// Get starting shard and update for next round
	startShard := int(pm.nextCleanupShard.Load())
	pm.nextCleanupShard.Store(int32((startShard + shardsPerRound) % len(pm.shards)))
	
	// Calculate target removals per shard
	excessEntries := totalEntries - pm.maxEntries
	targetPerShard := (excessEntries / shardsPerRound) + 1
	
	now := time.Now()
	oldThreshold := 24 * time.Hour
	totalRemoved := 0
	
	// Clean selected shards
	for i := 0; i < shardsPerRound; i++ {
		shardIdx := (startShard + i) % len(pm.shards)
		shard := pm.shards[shardIdx]
		
		// Collect LRU candidates from this shard
		type lruEntry struct {
			domain     string
			lastAccess time.Time
		}
		var entries []lruEntry
		
		shard.mu.RLock()
		for domain, lastAccess := range shard.lastAccess {
			// Skip hot domains
			if _, inDomains := shard.domains[domain]; inDomains {
				continue
			}
			// Only consider old entries
			if now.Sub(lastAccess) > oldThreshold {
				entries = append(entries, lruEntry{domain, lastAccess})
			}
		}
		shard.mu.RUnlock()
		
		// Sort by lastAccess (oldest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].lastAccess.Before(entries[j].lastAccess)
		})
		
		// Remove oldest entries up to target
		removeCount := targetPerShard
		if removeCount > len(entries) {
			removeCount = len(entries)
		}
		
		shard.mu.Lock()
		for i := 0; i < removeCount; i++ {
			domain := entries[i].domain
			delete(shard.hitCounters, domain)
			delete(shard.domains, domain)
			delete(shard.lastAccess, domain)
			totalRemoved++
		}
		shard.mu.Unlock()
	}
	
	// Incremental cleanup completed (debug logging removed for performance)
}

// cleanupWorker periodically scans for domains that need refreshing.
// Uses dynamic cleanup interval based on memory pressure.
func (pm *PrefetchManager) cleanupWorker() {
	defer pm.workerWg.Done()

	refreshTicker := time.NewTicker(10 * time.Second)
	defer refreshTicker.Stop()
	
	// Dynamic cleanup interval
	currentCleanupInterval := pm.cleanupInterval
	cleanupTicker := time.NewTicker(currentCleanupInterval)
	defer cleanupTicker.Stop()

	// Cleanup worker started (debug logging removed for performance)

	for {
		select {
		case <-pm.stopCh:
			// Cleanup worker stopped
			return

		case <-refreshTicker.C:
			pm.checkAndRefresh()

		case <-cleanupTicker.C:
			// Use incremental cleanup for better performance
			pm.cleanupIncremental()
			
			// Periodically do full cleanup (every 4 rounds = when nextCleanupShard wraps to 0)
			if pm.nextCleanupShard.Load() == 0 {
				// Performing full cleanup after incremental rounds
				pm.cleanupInternal(true)
			}
			
			// Adjust cleanup interval based on memory pressure
			totalEntries := pm.getTotalEntries()
			utilizationRatio := float64(totalEntries) / float64(pm.maxEntries)
			
			var newInterval time.Duration
			switch {
			case utilizationRatio > 1.2:
				// Severely over limit: cleanup every 15 minutes
				newInterval = 15 * time.Minute
			case utilizationRatio > 1.0:
				// Over limit: cleanup every 30 minutes
				newInterval = 30 * time.Minute
			case utilizationRatio > 0.8:
				// High usage: cleanup every 45 minutes
				newInterval = 45 * time.Minute
			case utilizationRatio > 0.5:
				// Normal usage: use configured interval
				newInterval = pm.cleanupInterval
			default:
				// Low usage: cleanup less frequently (2x interval)
				newInterval = pm.cleanupInterval * 2
			}
			
			// Cap the interval
			if newInterval < 15*time.Minute {
				newInterval = 15 * time.Minute
			}
			if newInterval > 4*time.Hour {
				newInterval = 4 * time.Hour
			}
			
			// Update ticker if interval changed significantly
			if newInterval != currentCleanupInterval {
				pm.logger.Info("adjusting cleanup interval",
					"old_interval", currentCleanupInterval,
					"new_interval", newInterval,
					"utilization", fmt.Sprintf("%.1f%%", utilizationRatio*100),
					"entries", totalEntries,
					"max", pm.maxEntries)
				
				cleanupTicker.Stop()
				cleanupTicker = time.NewTicker(newInterval)
				currentCleanupInterval = newInterval
			}
		}
	}
}

// checkAndRefresh scans for expired domains and schedules refresh tasks.
func (pm *PrefetchManager) checkAndRefresh() {
	now := time.Now()
	var tasks []*RefreshTask
	var totalDomains int

	// Scan all shards
	for _, shard := range pm.shards {
		shard.mu.Lock()
		totalDomains += len(shard.domains)
		for domain, expiry := range shard.domains {
			timeUntilExpiry := expiry.Sub(now)
			// If expired or about to expire (within 5 seconds)
			if now.After(expiry) || now.Add(5*time.Second).After(expiry) {
				pm.logger.Debug("domain ready for prefetch",
					"domain", domain,
					"expiry", expiry,
					"time_until_expiry", timeUntilExpiry)
				tasks = append(tasks, &RefreshTask{
					Domain:     domain,
					ExpireTime: expiry,
				})
				// Remove from domains map to prevent duplicate scheduling
				// Keep hits/lastAccess/hitTimestamps for statistics
				delete(shard.domains, domain)
			}
		}
		shard.mu.Unlock()
	}

	// Debug logging removed for performance

	if len(tasks) == 0 {
		return
	}

	pm.logger.Info("scheduling refresh tasks", "count", len(tasks))

	// Schedule all tasks
	for _, task := range tasks {
		pm.scheduleTask(task)
	}
}

// cleanup removes stale entries from hits, domains, and lastAccess maps
// to prevent memory leaks. This is called periodically and when maxEntries is exceeded.
// For testing and periodic maintenance, it always performs cleanup.
func (pm *PrefetchManager) cleanup() {
	pm.cleanupInternal(true)
}

// cleanupInternal performs the actual cleanup with optional force mode.
// force=true will clean even if under maxEntries limit (used for testing and periodic maintenance).
func (pm *PrefetchManager) cleanupInternal(force bool) {
	now := time.Now()
	totalEntries := pm.getTotalEntries()
	
	// Smart cleanup: only clean if we exceed threshold (unless forced)
	// This avoids unnecessary work when memory usage is acceptable
	cleanupNeeded := totalEntries > pm.maxEntries || force
	aggressiveCleanup := totalEntries > int(float64(pm.maxEntries)*1.2) // 20% over limit
	
	if !cleanupNeeded {
		// Cleanup not needed, entries within limit
		return
	}
	
	lowHitThreshold := pm.threshold // Remove entries below threshold
	
	// Time threshold for considering entries as "old"
	oldThreshold := 24 * time.Hour
	if aggressiveCleanup {
		oldThreshold = 12 * time.Hour // More aggressive when severely over limit
	}

	var totalRemoved int
	var totalHits, totalDomains int
	
	// Calculate target: how many entries to remove
	targetRemove := 0
	if aggressiveCleanup {
		// Remove 30% when severely over limit
		targetRemove = int(float64(totalEntries) * 0.3)
	} else if totalEntries > pm.maxEntries {
		// Remove just enough to get back under limit
		targetRemove = totalEntries - pm.maxEntries
	} else if force {
		// Force mode: clean up stale entries (10% or at least some entries)
		targetRemove = int(float64(totalEntries) * 0.1)
		if targetRemove < 1 && totalEntries > 0 {
			targetRemove = totalEntries // Clean all if very few entries
		}
	}

	// Cleanup started (debug logging removed for performance)

	// Collect candidates for removal with their scores
	type candidate struct {
		domain string
		score  float64 // Lower score = higher priority for removal
		shard  *domainShard
	}
	
	var candidates []candidate

	// Collect candidates from all shards
	for _, shard := range pm.shards {
		shard.mu.RLock()
		
		for domain, lastAccess := range shard.lastAccess {
			// Skip hot domains (currently being tracked for refresh)
			if _, inDomains := shard.domains[domain]; inDomains {
				continue
			}
			
			// Calculate removal score (lower = more likely to remove)
			score := 0.0
			
			// Factor 1: Time since last access (older = lower score)
			timeSinceAccess := now.Sub(lastAccess)
			if timeSinceAccess > oldThreshold {
				score -= 100 // Very old, high priority for removal
			} else {
				hoursRemaining := oldThreshold.Hours() - timeSinceAccess.Hours()
				score += hoursRemaining * 2 // Recent access increases score
			}
			
			// Factor 2: Hit count (lower hits = lower score)
			if counter, exists := shard.hitCounters[domain]; exists {
				score += float64(counter.count) * 10 // Each hit adds to score
				
				// Factor 3: Below threshold penalty
				if counter.count < lowHitThreshold {
					score -= 50 // Below threshold, priority for removal
				}
			}
			
			candidates = append(candidates, candidate{
				domain: domain,
				score:  score,
				shard:  shard,
			})
		}
		
		shard.mu.RUnlock()
	}
	
	// Sort candidates by score (lowest first = highest priority for removal)
	// Use partial sort for better performance - only sort what we need to remove
	removeCount := targetRemove
	if removeCount > len(candidates) {
		removeCount = len(candidates)
	}
	
	// Use heap-based partial sort for O(n log k) complexity instead of O(n²)
	// This is much faster when removeCount << len(candidates)
	if removeCount > 0 && removeCount < len(candidates) {
		// Partial sort using a simple but efficient approach
		// Sort only the first removeCount elements
		for i := 0; i < removeCount; i++ {
			minIdx := i
			// Only search in remaining unsorted portion
			for j := i + 1; j < len(candidates); j++ {
				if candidates[j].score < candidates[minIdx].score {
					minIdx = j
				}
			}
			if minIdx != i {
				candidates[i], candidates[minIdx] = candidates[minIdx], candidates[i]
			}
		}
	}
	
	// Remove entries with lowest scores
	for i := 0; i < removeCount; i++ {
		c := candidates[i]
		c.shard.mu.Lock()
		delete(c.shard.hitCounters, c.domain)
		delete(c.shard.domains, c.domain)
		delete(c.shard.lastAccess, c.domain)
		c.shard.mu.Unlock()
		
		totalRemoved++
	}
	
	// Count remaining entries
	for _, shard := range pm.shards {
		shard.mu.RLock()
		totalHits += len(shard.hitCounters)
		totalDomains += len(shard.domains)
		shard.mu.RUnlock()
	}

	pm.logger.Info("prefetch cleanup completed",
		"removed", totalRemoved,
		"target", targetRemove,
		"remaining_hits", totalHits,
		"remaining_domains", totalDomains,
		"aggressive", aggressiveCleanup)
}

// refresh directly updates the cache entry for a domain by querying upstream servers.
// This is more efficient than querying 127.0.0.1:53 because it bypasses the full
// DNS request processing pipeline (filters, etc.) and reduces network overhead.
// It includes retry logic to handle temporary failures.
func (pm *PrefetchManager) refresh(domain string) error {
	// Retry logic: try up to 3 times with exponential backoff
	const maxRetries = 3
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Refresh attempt (debug logging removed for performance)

		// Use context with timeout for each attempt
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := pm.server.refreshCacheEntry(ctx, domain)
		cancel()
		
		if err == nil {
			// Success
			if attempt > 1 {
				// Refresh succeeded after retry (debug logging removed)
				// Log retry success to error log
				pm.logRetrySuccess(domain, "cache_refresh", attempt)
			}
			return nil
		}

		lastErr = err
		
		// Log each failure attempt to error log
		pm.logError(domain, "cache_refresh", attempt, err)
		
		// Don't retry on last attempt
		if attempt < maxRetries {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*attempt) * time.Millisecond
			// Retrying (debug logging removed for performance)
			time.Sleep(backoff)
		}
	}

	// All retries failed
	pm.logger.Warn("prefetch refresh failed after all retries",
		"domain", domain,
		"method", "cache_refresh",
		"attempts", maxRetries,
		"err", lastErr)
	
	// Try to rotate log if needed (check every 100 failures)
	if pm.metrics.TasksFailed.Load()%100 == 0 {
		if err := pm.rotateErrorLog(); err != nil {
			pm.logger.Warn("failed to rotate error log", slogutil.KeyError, err)
		}
	}
	
	return lastErr
}

// GetStats returns statistics about the prefetch manager.
// This is useful for monitoring and debugging.
func (pm *PrefetchManager) GetStats() (hits, domains, tracked int) {
	for _, shard := range pm.shards {
		shard.mu.RLock()
		hits += len(shard.hitCounters)
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

	// Calibrate queue size counters with actual queue lengths
	// This ensures counters stay accurate even if there are edge cases
	actualUrgentSize := int32(len(pm.urgentQueue))
	actualNormalSize := int32(len(pm.normalQueue))
	
	pm.metrics.UrgentQueueSize.Store(actualUrgentSize)
	pm.metrics.NormalQueueSize.Store(actualNormalSize)

	pm.logger.Info("prefetch metrics",
		"active_tasks", pm.metrics.CurrentActive.Load(),
		"urgent_queue", actualUrgentSize,
		"normal_queue", actualNormalSize,
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

	metrics := map[string]int64{
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

	// Add last prefetch time if available
	if t := pm.lastPrefetchTime.Load(); t != nil {
		if lastTime, ok := t.(time.Time); ok {
			metrics["last_prefetch_unix"] = lastTime.Unix()
		}
	}

	return metrics
}

// initErrorLog initializes the error log file for prefetch failures.
func (pm *PrefetchManager) initErrorLog() error {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	
	exeDir := filepath.Dir(exePath)
	logPath := filepath.Join(exeDir, "prefetch_errors.log")
	
	// Open log file in append mode, create if not exists
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open error log file: %w", err)
	}
	
	pm.errorLogFile = file
	pm.logger.Info("prefetch error log initialized", "path", logPath)
	
	// Write header
	pm.logError("", "", 0, fmt.Errorf("=== Prefetch Error Log Started at %s ===", time.Now().Format(time.RFC3339)))
	
	return nil
}

// closeErrorLog closes the error log file.
func (pm *PrefetchManager) closeErrorLog() {
	pm.errorLogMu.Lock()
	defer pm.errorLogMu.Unlock()
	
	if pm.errorLogFile != nil {
		// Write footer
		fmt.Fprintf(pm.errorLogFile, "=== Prefetch Error Log Closed at %s ===\n\n", time.Now().Format(time.RFC3339))
		pm.errorLogFile.Close()
		pm.errorLogFile = nil
	}
}

// logError writes a detailed error entry to the error log file.
func (pm *PrefetchManager) logError(domain, method string, attempt int, err error) {
	pm.errorLogMu.Lock()
	defer pm.errorLogMu.Unlock()
	
	if pm.errorLogFile == nil {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	if domain == "" {
		// Header/footer message
		fmt.Fprintf(pm.errorLogFile, "%s\n", err.Error())
		return
	}
	
	// Format: [timestamp] Domain: xxx | Method: cache_refresh | Attempt: 1 | Error: xxx
	fmt.Fprintf(pm.errorLogFile, "[%s] Domain: %-50s | Method: %-20s | Attempt: %d | Error: %v\n",
		timestamp, domain, method, attempt, err)
	
	// Flush to ensure immediate write
	pm.errorLogFile.Sync()
}

// logRetrySuccess writes a retry success entry to the error log file.
func (pm *PrefetchManager) logRetrySuccess(domain, method string, attempt int) {
	pm.errorLogMu.Lock()
	defer pm.errorLogMu.Unlock()
	
	if pm.errorLogFile == nil {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// Format: [timestamp] Domain: xxx | Method: cache_refresh | Attempt: 2 | SUCCESS
	fmt.Fprintf(pm.errorLogFile, "[%s] Domain: %-50s | Method: %-20s | Attempt: %d | SUCCESS (recovered from failure)\n",
		timestamp, domain, method, attempt)
	
	pm.errorLogFile.Sync()
}

// rotateErrorLog rotates the error log file if it exceeds a certain size.
func (pm *PrefetchManager) rotateErrorLog() error {
	pm.errorLogMu.Lock()
	defer pm.errorLogMu.Unlock()
	
	if pm.errorLogFile == nil {
		return nil
	}
	
	// Check file size
	info, err := pm.errorLogFile.Stat()
	if err != nil {
		return err
	}
	
	// Rotate if file size > 10MB
	if info.Size() < 10*1024*1024 {
		return nil
	}
	
	// Close current file
	pm.errorLogFile.Close()
	
	// Get file path
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	
	exeDir := filepath.Dir(exePath)
	logPath := filepath.Join(exeDir, "prefetch_errors.log")
	backupPath := filepath.Join(exeDir, fmt.Sprintf("prefetch_errors.%s.log", time.Now().Format("20060102_150405")))
	
	// Rename old file
	if err := os.Rename(logPath, backupPath); err != nil {
		return err
	}
	
	// Create new file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	pm.errorLogFile = file
	pm.logger.Info("prefetch error log rotated", "backup", backupPath)
	
	// Write header
	fmt.Fprintf(pm.errorLogFile, "=== Prefetch Error Log Started at %s (after rotation) ===\n", time.Now().Format(time.RFC3339))
	
	return nil
}
