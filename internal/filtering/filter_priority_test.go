package filtering

import (
	"context"
	"testing"

	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/stretchr/testify/assert"
)

func TestDetectPriorityConflicts(t *testing.T) {
	// Create a test logger
	logger := slogutil.NewDiscardLogger()

	d := &DNSFilter{
		logger: logger,
	}

	ctx := context.Background()

	tests := []struct {
		name            string
		filters         []FilterYAML
		expectConflicts bool
		conflictCount   int
	}{
		{
			name:            "no_filters",
			filters:         []FilterYAML{},
			expectConflicts: false,
		},
		{
			name: "single_filter",
			filters: []FilterYAML{
				{Name: "Filter1", Priority: 10},
			},
			expectConflicts: false,
		},
		{
			name: "no_conflicts",
			filters: []FilterYAML{
				{Name: "Filter1", Priority: 10},
				{Name: "Filter2", Priority: 20},
				{Name: "Filter3", Priority: 30},
			},
			expectConflicts: false,
		},
		{
			name: "two_filters_same_priority",
			filters: []FilterYAML{
				{Name: "Filter1", Priority: 10},
				{Name: "Filter2", Priority: 10},
			},
			expectConflicts: true,
			conflictCount:   1,
		},
		{
			name: "three_filters_same_priority",
			filters: []FilterYAML{
				{Name: "Filter1", Priority: 10},
				{Name: "Filter2", Priority: 10},
				{Name: "Filter3", Priority: 10},
			},
			expectConflicts: true,
			conflictCount:   1,
		},
		{
			name: "multiple_conflict_groups",
			filters: []FilterYAML{
				{Name: "Filter1", Priority: 10},
				{Name: "Filter2", Priority: 10},
				{Name: "Filter3", Priority: 20},
				{Name: "Filter4", Priority: 20},
				{Name: "Filter5", Priority: 30},
			},
			expectConflicts: true,
			conflictCount:   2,
		},
		{
			name: "mixed_conflicts_and_unique",
			filters: []FilterYAML{
				{Name: "Filter1", Priority: 10},
				{Name: "Filter2", Priority: 10},
				{Name: "Filter3", Priority: 20},
				{Name: "Filter4", Priority: 30},
				{Name: "Filter5", Priority: 30},
			},
			expectConflicts: true,
			conflictCount:   2,
		},
		{
			name: "filters_without_names",
			filters: []FilterYAML{
				{URL: "http://example.com/1", Priority: 10},
				{URL: "http://example.com/2", Priority: 10},
			},
			expectConflicts: true,
			conflictCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the function doesn't panic
			assert.NotPanics(t, func() {
				d.detectPriorityConflicts(ctx, tt.filters)
			})
		})
	}
}

func TestDetectPriorityConflicts_EdgeCases(t *testing.T) {
	logger := slogutil.NewDiscardLogger()

	d := &DNSFilter{
		logger: logger,
	}

	ctx := context.Background()

	t.Run("nil_filters", func(t *testing.T) {
		assert.NotPanics(t, func() {
			d.detectPriorityConflicts(ctx, nil)
		})
	})

	t.Run("all_same_priority", func(t *testing.T) {
		filters := []FilterYAML{
			{Name: "Filter1", Priority: 50},
			{Name: "Filter2", Priority: 50},
			{Name: "Filter3", Priority: 50},
			{Name: "Filter4", Priority: 50},
			{Name: "Filter5", Priority: 50},
		}
		
		assert.NotPanics(t, func() {
			d.detectPriorityConflicts(ctx, filters)
		})
	})

	t.Run("priority_boundaries", func(t *testing.T) {
		filters := []FilterYAML{
			{Name: "Filter1", Priority: 0},
			{Name: "Filter2", Priority: 0},
			{Name: "Filter3", Priority: 100},
			{Name: "Filter4", Priority: 100},
		}
		
		assert.NotPanics(t, func() {
			d.detectPriorityConflicts(ctx, filters)
		})
	})
}

// Benchmark test
func BenchmarkDetectPriorityConflicts(b *testing.B) {
	logger := slogutil.NewDiscardLogger()

	d := &DNSFilter{
		logger: logger,
	}

	ctx := context.Background()

	// Create test data with some conflicts
	filters := []FilterYAML{
		{Name: "Filter1", Priority: 10},
		{Name: "Filter2", Priority: 10},
		{Name: "Filter3", Priority: 20},
		{Name: "Filter4", Priority: 30},
		{Name: "Filter5", Priority: 30},
		{Name: "Filter6", Priority: 40},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.detectPriorityConflicts(ctx, filters)
	}
}

func BenchmarkDetectPriorityConflicts_NoConflicts(b *testing.B) {
	logger := slogutil.NewDiscardLogger()

	d := &DNSFilter{
		logger: logger,
	}

	ctx := context.Background()

	// Create test data without conflicts
	filters := make([]FilterYAML, 100)
	for i := 0; i < 100; i++ {
		filters[i] = FilterYAML{
			Name:     "Filter" + string(rune(i)),
			Priority: i,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.detectPriorityConflicts(ctx, filters)
	}
}
