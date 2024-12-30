package translationflow_test

import (
	"testing"

	"raggo/src/translationflow"
)

func TestCalculateChunkSize(t *testing.T) {
	tests := []struct {
		name       string
		tokenCount int
		tokenLimit int
		wantSize   int
	}{
		{
			name:       "below limit",
			tokenCount: 1000,
			tokenLimit: 500,
			wantSize:   500,
		},
		{
			name:       "above limit - case 1",
			tokenCount: 1530,
			tokenLimit: 500,
			wantSize:   389,
		},
		{
			name:       "above limit - case 2",
			tokenCount: 2242,
			tokenLimit: 500,
			wantSize:   496,
		},
		{
			name:       "equal to limit",
			tokenCount: 500,
			tokenLimit: 500,
			wantSize:   500,
		},
		{
			name:       "small numbers",
			tokenCount: 10,
			tokenLimit: 20,
			wantSize:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translationflow.CalculateChunkSize(tt.tokenCount, tt.tokenLimit)
			if got != tt.wantSize {
				t.Errorf("CalculateChunkSize(%d, %d) = %d, want %d",
					tt.tokenCount, tt.tokenLimit, got, tt.wantSize)
			}
		})
	}
}
