package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid_priority_0",
			priority: 0,
			wantErr:  false,
		},
		{
			name:     "valid_priority_50",
			priority: 50,
			wantErr:  false,
		},
		{
			name:     "valid_priority_100",
			priority: 100,
			wantErr:  false,
		},
		{
			name:     "valid_priority_1",
			priority: 1,
			wantErr:  false,
		},
		{
			name:     "valid_priority_99",
			priority: 99,
			wantErr:  false,
		},
		{
			name:     "invalid_priority_negative",
			priority: -1,
			wantErr:  true,
			errMsg:   "priority must be between 0 and 100",
		},
		{
			name:     "invalid_priority_too_high",
			priority: 101,
			wantErr:  true,
			errMsg:   "priority must be between 0 and 100",
		},
		{
			name:     "invalid_priority_very_negative",
			priority: -100,
			wantErr:  true,
			errMsg:   "priority must be between 0 and 100",
		},
		{
			name:     "invalid_priority_very_high",
			priority: 1000,
			wantErr:  true,
			errMsg:   "priority must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePriority(tt.priority)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark test
func BenchmarkValidatePriority(b *testing.B) {
	priority := 50
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validatePriority(priority)
	}
}
