package dnsrouting

import (
	"context"
	"sync"
)

// BatchQuery represents a single query in a batch
type BatchQuery struct {
	Domain string
	Index  int // Original index in the batch
}

// BatchResult represents the result of a batch query
type BatchResult struct {
	Domain        string
	UpstreamGroup string
	Matched       bool
	Index         int
	Error         error
}

// BatchMatcher provides batch query functionality
type BatchMatcher struct {
	router      *Router
	workerCount int
}

// NewBatchMatcher creates a new batch matcher
func NewBatchMatcher(router *Router, workerCount int) *BatchMatcher {
	if workerCount <= 0 {
		workerCount = 4 // Default worker count
	}
	
	return &BatchMatcher{
		router:      router,
		workerCount: workerCount,
	}
}

// MatchBatch performs batch matching for multiple domains
// Returns results in the same order as input domains
func (b *BatchMatcher) MatchBatch(ctx context.Context, domains []string) []BatchResult {
	if len(domains) == 0 {
		return []BatchResult{}
	}

	// Create result slice with same length as input
	results := make([]BatchResult, len(domains))
	
	// Create work queue
	queries := make(chan BatchQuery, len(domains))
	resultsChan := make(chan BatchResult, len(domains))
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < b.workerCount; i++ {
		wg.Add(1)
		go b.worker(ctx, queries, resultsChan, &wg)
	}
	
	// Send queries to workers
	go func() {
		for i, domain := range domains {
			select {
			case <-ctx.Done():
				return
			case queries <- BatchQuery{Domain: domain, Index: i}:
			}
		}
		close(queries)
	}()
	
	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Gather results
	for result := range resultsChan {
		if result.Index >= 0 && result.Index < len(results) {
			results[result.Index] = result
		}
	}
	
	return results
}

// worker processes queries from the queue
func (b *BatchMatcher) worker(ctx context.Context, queries <-chan BatchQuery, results chan<- BatchResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for query := range queries {
		select {
		case <-ctx.Done():
			results <- BatchResult{
				Domain: query.Domain,
				Index:  query.Index,
				Error:  ctx.Err(),
			}
			return
		default:
			// Perform the match
			upstreamGroup, matched := b.router.Match(ctx, query.Domain)
			
			results <- BatchResult{
				Domain:        query.Domain,
				UpstreamGroup: upstreamGroup,
				Matched:       matched,
				Index:         query.Index,
			}
		}
	}
}

// MatchBatchSimple performs batch matching without workers (simpler but slower)
func (b *BatchMatcher) MatchBatchSimple(ctx context.Context, domains []string) []BatchResult {
	results := make([]BatchResult, len(domains))
	
	for i, domain := range domains {
		select {
		case <-ctx.Done():
			results[i] = BatchResult{
				Domain: domain,
				Index:  i,
				Error:  ctx.Err(),
			}
			return results
		default:
			upstreamGroup, matched := b.router.Match(ctx, domain)
			results[i] = BatchResult{
				Domain:        domain,
				UpstreamGroup: upstreamGroup,
				Matched:       matched,
				Index:         i,
			}
		}
	}
	
	return results
}

// MatchBatchMap performs batch matching and returns a map
// Useful when order doesn't matter
func (b *BatchMatcher) MatchBatchMap(ctx context.Context, domains []string) map[string]BatchResult {
	results := b.MatchBatch(ctx, domains)
	
	resultMap := make(map[string]BatchResult, len(results))
	for _, result := range results {
		resultMap[result.Domain] = result
	}
	
	return resultMap
}

// BatchMatchWithCallback performs batch matching with a callback for each result
// Useful for streaming results
func (b *BatchMatcher) BatchMatchWithCallback(ctx context.Context, domains []string, callback func(BatchResult)) {
	if len(domains) == 0 {
		return
	}

	queries := make(chan BatchQuery, len(domains))
	results := make(chan BatchResult, b.workerCount)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < b.workerCount; i++ {
		wg.Add(1)
		go b.worker(ctx, queries, results, &wg)
	}
	
	// Send queries
	go func() {
		for i, domain := range domains {
			select {
			case <-ctx.Done():
				return
			case queries <- BatchQuery{Domain: domain, Index: i}:
			}
		}
		close(queries)
	}()
	
	// Process results with callback
	go func() {
		wg.Wait()
		close(results)
	}()
	
	for result := range results {
		callback(result)
	}
}

// Stats returns batch matcher statistics
type BatchStats struct {
	WorkerCount int
	QueueSize   int
}

// GetStats returns current batch matcher statistics
func (b *BatchMatcher) GetStats() BatchStats {
	return BatchStats{
		WorkerCount: b.workerCount,
	}
}
