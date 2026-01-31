package layout

import (
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

func TestMinY(t *testing.T) {
	tests := []struct {
		name     string
		elements []TextElement
		want     float64
	}{
		{
			name:     "empty slice",
			elements: nil,
			want:     0,
		},
		{
			name: "single element",
			elements: []TextElement{
				{Y: 100},
			},
			want: 100,
		},
		{
			name: "multiple elements",
			elements: []TextElement{
				{Y: 100},
				{Y: 50},
				{Y: 150},
			},
			want: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinY(tt.elements)
			if got != tt.want {
				t.Errorf("MinY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxY(t *testing.T) {
	tests := []struct {
		name     string
		elements []TextElement
		want     float64
	}{
		{
			name:     "empty slice",
			elements: nil,
			want:     0,
		},
		{
			name: "single element",
			elements: []TextElement{
				{Y: 100, Height: 10},
			},
			want: 110,
		},
		{
			name: "multiple elements",
			elements: []TextElement{
				{Y: 100, Height: 10},
				{Y: 50, Height: 20},
				{Y: 150, Height: 5},
			},
			want: 155,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxY(tt.elements)
			if got != tt.want {
				t.Errorf("MaxY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinX(t *testing.T) {
	tests := []struct {
		name     string
		elements []TextElement
		want     float64
	}{
		{
			name:     "empty slice",
			elements: nil,
			want:     0,
		},
		{
			name: "multiple elements",
			elements: []TextElement{
				{X: 100},
				{X: 50},
				{X: 150},
			},
			want: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinX(tt.elements)
			if got != tt.want {
				t.Errorf("MinX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAvgFontSize(t *testing.T) {
	tests := []struct {
		name     string
		elements []TextElement
		want     float64
	}{
		{
			name:     "empty slice",
			elements: nil,
			want:     0,
		},
		{
			name: "single element",
			elements: []TextElement{
				{Size: 12},
			},
			want: 12,
		},
		{
			name: "multiple elements",
			elements: []TextElement{
				{Size: 10},
				{Size: 12},
				{Size: 14},
			},
			want: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AvgFontSize(tt.elements)
			if got != tt.want {
				t.Errorf("AvgFontSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name string
		obj  core.Object
		want float64
	}{
		{
			name: "integer",
			obj:  core.Integer(42),
			want: 42,
		},
		{
			name: "real",
			obj:  core.Real(3.14),
			want: 3.14,
		},
		{
			name: "string returns 0",
			obj:  core.String("hello"),
			want: 0,
		},
		{
			name: "nil returns 0",
			obj:  nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToFloat64(tt.obj)
			if got != tt.want {
				t.Errorf("ToFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}
