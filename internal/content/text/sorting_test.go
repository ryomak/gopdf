package text

import (
	"testing"
)

// testElement implements SortableTextElement for testing.
type testElement struct {
	X, Y, Size float64
}

func (e testElement) GetX() float64    { return e.X }
func (e testElement) GetY() float64    { return e.Y }
func (e testElement) GetSize() float64 { return e.Size }

func TestGroupByLine(t *testing.T) {
	tests := []struct {
		name         string
		elements     []testElement
		wantNumLines int
	}{
		{
			name:         "empty slice",
			elements:     nil,
			wantNumLines: 0,
		},
		{
			name: "single element",
			elements: []testElement{
				{X: 0, Y: 100, Size: 12},
			},
			wantNumLines: 1,
		},
		{
			name: "same line elements",
			elements: []testElement{
				{X: 0, Y: 100, Size: 12},
				{X: 50, Y: 100, Size: 12},
				{X: 100, Y: 100, Size: 12},
			},
			wantNumLines: 1,
		},
		{
			name: "different lines",
			elements: []testElement{
				{X: 0, Y: 100, Size: 12},
				{X: 0, Y: 80, Size: 12},
				{X: 0, Y: 60, Size: 12},
			},
			wantNumLines: 3,
		},
		{
			name: "mixed same and different lines",
			elements: []testElement{
				{X: 0, Y: 100, Size: 12},
				{X: 50, Y: 100, Size: 12},
				{X: 0, Y: 80, Size: 12},
			},
			wantNumLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupByLine(tt.elements)
			if len(got) != tt.wantNumLines {
				t.Errorf("GroupByLine() returned %d lines, want %d", len(got), tt.wantNumLines)
			}
		})
	}
}

func TestSortByReadingOrder(t *testing.T) {
	tests := []struct {
		name     string
		elements []testElement
		wantX    []float64
	}{
		{
			name:     "empty slice",
			elements: nil,
			wantX:    nil,
		},
		{
			name: "already sorted",
			elements: []testElement{
				{X: 0, Y: 100, Size: 12},
				{X: 50, Y: 100, Size: 12},
				{X: 0, Y: 80, Size: 12},
			},
			wantX: []float64{0, 50, 0},
		},
		{
			name: "needs sorting",
			elements: []testElement{
				{X: 50, Y: 100, Size: 12},
				{X: 0, Y: 80, Size: 12},
				{X: 0, Y: 100, Size: 12},
			},
			wantX: []float64{0, 50, 0}, // First line: X=0, X=50; Second line: X=0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortByReadingOrder(tt.elements)
			if tt.wantX == nil {
				if len(got) != 0 {
					t.Errorf("SortByReadingOrder() = %v, want empty", got)
				}
				return
			}

			if len(got) != len(tt.wantX) {
				t.Errorf("SortByReadingOrder() length = %d, want %d", len(got), len(tt.wantX))
				return
			}

			for i, x := range tt.wantX {
				if got[i].GetX() != x {
					t.Errorf("SortByReadingOrder()[%d].X = %v, want %v", i, got[i].GetX(), x)
				}
			}
		})
	}
}
