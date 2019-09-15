package rat

import (
	"reflect"
	"testing"
)

func TestEvenSizes(t *testing.T) {
	tests := []struct {
		n         int
		totalSize int
		expected  []section
	}{
		{1, 120, []section{section{0, 120}}},
		{2, 120, []section{section{0, 59}, section{60, 60}}},
		{3, 120, []section{section{0, 39}, section{40, 39}, section{80, 40}}},
	}

	for _, tt := range tests {
		sections := evenSizes(tt.n, tt.totalSize)
		if !reflect.DeepEqual(sections, tt.expected) {
			t.Errorf("evenSizes(%d, %d): expected %v, got %v", tt.n, tt.totalSize, tt.expected, sections)
		}
	}
}

func TestGoldenSizes(t *testing.T) {
	tests := []struct {
		n         int
		totalSize int
		expected  []section
	}{
		{1, 120, []section{section{0, 120}}},
		{2, 120, []section{section{0, 45}, section{46, 74}}},
		{3, 120, []section{section{0, 17}, section{18, 27}, section{46, 74}}},
	}

	for _, tt := range tests {
		sections := goldenSizes(tt.n, tt.totalSize)
		if !reflect.DeepEqual(sections, tt.expected) {
			t.Errorf("goldenSizes(%d, %d): expected %v, got %v", tt.n, tt.totalSize, tt.expected, sections)
		}
	}
}
