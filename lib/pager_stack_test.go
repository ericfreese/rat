package rat

import (
	"reflect"
	"testing"
)

func TestEvenSizes(t *testing.T) {
	tests := []struct {
		n         int
		totalSize int
		expected  []boxSize
	}{
		{1, 120, []boxSize{boxSize{0, 120}}},
		{2, 120, []boxSize{boxSize{0, 59}, boxSize{60, 60}}},
		{3, 120, []boxSize{boxSize{0, 39}, boxSize{40, 39}, boxSize{80, 40}}},
	}

	for _, tt := range tests {
		boxes := evenSizes(tt.n, tt.totalSize)
		if !reflect.DeepEqual(boxes, tt.expected) {
			t.Errorf("evenSizes(%d, %d): expected %v, got %v", tt.n, tt.totalSize, tt.expected, boxes)
		}
	}
}

func TestGoldenSizes(t *testing.T) {
	tests := []struct {
		n         int
		totalSize int
		expected  []boxSize
	}{
		{1, 120, []boxSize{boxSize{0, 120}}},
		{2, 120, []boxSize{boxSize{0, 45}, boxSize{46, 74}}},
		{3, 120, []boxSize{boxSize{0, 17}, boxSize{18, 27}, boxSize{46, 74}}},
	}

	for _, tt := range tests {
		boxes := goldenSizes(tt.n, tt.totalSize)
		if !reflect.DeepEqual(boxes, tt.expected) {
			t.Errorf("goldenSizes(%d, %d): expected %v, got %v", tt.n, tt.totalSize, tt.expected, boxes)
		}
	}
}
