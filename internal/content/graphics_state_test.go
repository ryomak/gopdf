package content

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestIdentity(t *testing.T) {
	m := Identity()
	if m.A != 1 || m.B != 0 || m.C != 0 || m.D != 1 || m.E != 0 || m.F != 0 {
		t.Errorf("Identity() = %+v, want {1 0 0 1 0 0}", m)
	}
}

func TestMatrix_Multiply(t *testing.T) {
	tests := []struct {
		name   string
		m1     Matrix
		m2     Matrix
		expect Matrix
	}{
		{
			name:   "identity * identity",
			m1:     Identity(),
			m2:     Identity(),
			expect: Identity(),
		},
		{
			name:   "scale 2x * scale 3x",
			m1:     Matrix{A: 2, B: 0, C: 0, D: 2, E: 0, F: 0},
			m2:     Matrix{A: 3, B: 0, C: 0, D: 3, E: 0, F: 0},
			expect: Matrix{A: 6, B: 0, C: 0, D: 6, E: 0, F: 0},
		},
		{
			name:   "identity * translation",
			m1:     Identity(),
			m2:     Matrix{A: 1, B: 0, C: 0, D: 1, E: 10, F: 20},
			expect: Matrix{A: 1, B: 0, C: 0, D: 1, E: 10, F: 20},
		},
		{
			name:   "translation * identity",
			m1:     Matrix{A: 1, B: 0, C: 0, D: 1, E: 10, F: 20},
			m2:     Identity(),
			expect: Matrix{A: 1, B: 0, C: 0, D: 1, E: 10, F: 20},
		},
		{
			name:   "scale * translation",
			m1:     Matrix{A: 2, B: 0, C: 0, D: 3, E: 0, F: 0},
			m2:     Matrix{A: 1, B: 0, C: 0, D: 1, E: 5, F: 10},
			expect: Matrix{A: 2, B: 0, C: 0, D: 3, E: 5, F: 10},
		},
		{
			name: "90 degree rotation",
			m1:   Identity(),
			m2:   Matrix{A: 0, B: 1, C: -1, D: 0, E: 0, F: 0},
			expect: Matrix{A: 0, B: 1, C: -1, D: 0, E: 0, F: 0},
		},
		{
			name: "two translations compose",
			m1:   Matrix{A: 1, B: 0, C: 0, D: 1, E: 10, F: 20},
			m2:   Matrix{A: 1, B: 0, C: 0, D: 1, E: 30, F: 40},
			expect: Matrix{A: 1, B: 0, C: 0, D: 1, E: 40, F: 60},
		},
		{
			name: "general matrices",
			m1:   Matrix{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6},
			m2:   Matrix{A: 7, B: 8, C: 9, D: 10, E: 11, F: 12},
			// A = 1*7 + 2*9 = 25, B = 1*8 + 2*10 = 28
			// C = 3*7 + 4*9 = 57, D = 3*8 + 4*10 = 64
			// E = 5*7 + 6*9 + 11 = 100, F = 5*8 + 6*10 + 12 = 112
			expect: Matrix{A: 25, B: 28, C: 57, D: 64, E: 100, F: 112},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m1.Multiply(tt.m2)
			if !approxEqual(got.A, tt.expect.A) || !approxEqual(got.B, tt.expect.B) ||
				!approxEqual(got.C, tt.expect.C) || !approxEqual(got.D, tt.expect.D) ||
				!approxEqual(got.E, tt.expect.E) || !approxEqual(got.F, tt.expect.F) {
				t.Errorf("Multiply() = %+v, want %+v", got, tt.expect)
			}
		})
	}
}

func TestMatrix_TransformPoint(t *testing.T) {
	tests := []struct {
		name       string
		m          Matrix
		x, y       float64
		expectX    float64
		expectY    float64
	}{
		{
			name:    "identity",
			m:       Identity(),
			x:       10, y: 20,
			expectX: 10, expectY: 20,
		},
		{
			name:    "scale 2x",
			m:       Matrix{A: 2, B: 0, C: 0, D: 2, E: 0, F: 0},
			x:       5, y: 10,
			expectX: 10, expectY: 20,
		},
		{
			name:    "translation",
			m:       Matrix{A: 1, B: 0, C: 0, D: 1, E: 100, F: 200},
			x:       10, y: 20,
			expectX: 110, expectY: 220,
		},
		{
			name:    "scale and translate",
			m:       Matrix{A: 2, B: 0, C: 0, D: 3, E: 10, F: 20},
			x:       5, y: 10,
			expectX: 20, expectY: 50,
		},
		{
			name:    "90 degree rotation",
			m:       Matrix{A: 0, B: 1, C: -1, D: 0, E: 0, F: 0},
			x:       10, y: 0,
			expectX: 0, expectY: 10,
		},
		{
			name:    "origin point with translation",
			m:       Matrix{A: 1, B: 0, C: 0, D: 1, E: 50, F: 60},
			x:       0, y: 0,
			expectX: 50, expectY: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := tt.m.TransformPoint(tt.x, tt.y)
			if !approxEqual(gotX, tt.expectX) || !approxEqual(gotY, tt.expectY) {
				t.Errorf("TransformPoint(%v, %v) = (%v, %v), want (%v, %v)",
					tt.x, tt.y, gotX, gotY, tt.expectX, tt.expectY)
			}
		})
	}
}

func TestMatrix_TransformRect(t *testing.T) {
	tests := []struct {
		name                       string
		m                          Matrix
		x, y, width, height        float64
		expectMinX, expectMinY     float64
		expectMaxX, expectMaxY     float64
	}{
		{
			name:       "identity",
			m:          Identity(),
			x:          10, y: 20, width: 100, height: 50,
			expectMinX: 10, expectMinY: 20,
			expectMaxX: 110, expectMaxY: 70,
		},
		{
			name:       "scale 2x",
			m:          Matrix{A: 2, B: 0, C: 0, D: 2, E: 0, F: 0},
			x:          10, y: 20, width: 100, height: 50,
			expectMinX: 20, expectMinY: 40,
			expectMaxX: 220, expectMaxY: 140,
		},
		{
			name:       "translation",
			m:          Matrix{A: 1, B: 0, C: 0, D: 1, E: 50, F: 100},
			x:          10, y: 20, width: 30, height: 40,
			expectMinX: 60, expectMinY: 120,
			expectMaxX: 90, expectMaxY: 160,
		},
		{
			name:       "90 degree rotation - rect becomes rotated bounding box",
			m:          Matrix{A: 0, B: 1, C: -1, D: 0, E: 0, F: 0},
			x:          0, y: 0, width: 100, height: 50,
			expectMinX: -50, expectMinY: 0,
			expectMaxX: 0, expectMaxY: 100,
		},
		{
			name:       "zero-size rect",
			m:          Identity(),
			x:          5, y: 10, width: 0, height: 0,
			expectMinX: 5, expectMinY: 10,
			expectMaxX: 5, expectMaxY: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minX, minY, maxX, maxY := tt.m.TransformRect(tt.x, tt.y, tt.width, tt.height)
			if !approxEqual(minX, tt.expectMinX) || !approxEqual(minY, tt.expectMinY) ||
				!approxEqual(maxX, tt.expectMaxX) || !approxEqual(maxY, tt.expectMaxY) {
				t.Errorf("TransformRect() = (%v, %v, %v, %v), want (%v, %v, %v, %v)",
					minX, minY, maxX, maxY,
					tt.expectMinX, tt.expectMinY, tt.expectMaxX, tt.expectMaxY)
			}
		})
	}
}

func TestMatrix_Inverse(t *testing.T) {
	tests := []struct {
		name   string
		m      Matrix
		expect Matrix
	}{
		{
			name:   "identity inverse is identity",
			m:      Identity(),
			expect: Identity(),
		},
		{
			name:   "scale 2x inverse is scale 0.5x",
			m:      Matrix{A: 2, B: 0, C: 0, D: 2, E: 0, F: 0},
			expect: Matrix{A: 0.5, B: 0, C: 0, D: 0.5, E: 0, F: 0},
		},
		{
			name:   "translation inverse negates translation",
			m:      Matrix{A: 1, B: 0, C: 0, D: 1, E: 10, F: 20},
			expect: Matrix{A: 1, B: 0, C: 0, D: 1, E: -10, F: -20},
		},
		{
			name:   "near-zero determinant returns identity",
			m:      Matrix{A: 0, B: 0, C: 0, D: 0, E: 5, F: 10},
			expect: Identity(),
		},
		{
			name:   "singular matrix (det=0) returns identity",
			m:      Matrix{A: 1, B: 2, C: 2, D: 4, E: 0, F: 0},
			expect: Identity(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Inverse()
			if !approxEqual(got.A, tt.expect.A) || !approxEqual(got.B, tt.expect.B) ||
				!approxEqual(got.C, tt.expect.C) || !approxEqual(got.D, tt.expect.D) ||
				!approxEqual(got.E, tt.expect.E) || !approxEqual(got.F, tt.expect.F) {
				t.Errorf("Inverse() = %+v, want %+v", got, tt.expect)
			}
		})
	}
}

func TestMatrix_Inverse_RoundTrip(t *testing.T) {
	// M * M^-1 should equal identity
	m := Matrix{A: 3, B: 1, C: 2, D: 4, E: 5, F: 6}
	inv := m.Inverse()
	result := m.Multiply(inv)
	id := Identity()

	if !approxEqual(result.A, id.A) || !approxEqual(result.B, id.B) ||
		!approxEqual(result.C, id.C) || !approxEqual(result.D, id.D) ||
		!approxEqual(result.E, id.E) || !approxEqual(result.F, id.F) {
		t.Errorf("M * M^-1 = %+v, want identity %+v", result, id)
	}
}

func TestNewGraphicsState(t *testing.T) {
	gs := NewGraphicsState()

	// CTM should be identity
	id := Identity()
	if gs.CTM != id {
		t.Errorf("CTM = %+v, want identity", gs.CTM)
	}

	// LastCMMatrix should be nil
	if gs.LastCMMatrix != nil {
		t.Errorf("LastCMMatrix should be nil, got %+v", gs.LastCMMatrix)
	}

	// ColorSpace
	if gs.ColorSpace != "DeviceRGB" {
		t.Errorf("ColorSpace = %q, want %q", gs.ColorSpace, "DeviceRGB")
	}

	// StrokeColor should be black
	if gs.StrokeColor != [3]float64{0, 0, 0} {
		t.Errorf("StrokeColor = %v, want [0 0 0]", gs.StrokeColor)
	}

	// FillColor should be black
	if gs.FillColor != [3]float64{0, 0, 0} {
		t.Errorf("FillColor = %v, want [0 0 0]", gs.FillColor)
	}

	// LineWidth
	if gs.LineWidth != 1.0 {
		t.Errorf("LineWidth = %v, want 1.0", gs.LineWidth)
	}
}

func TestGraphicsState_Clone(t *testing.T) {
	gs := NewGraphicsState()
	gs.CTM = Matrix{A: 2, B: 0, C: 0, D: 3, E: 10, F: 20}
	gs.ColorSpace = "DeviceCMYK"
	gs.StrokeColor = [3]float64{1, 0, 0}
	gs.FillColor = [3]float64{0, 1, 0}
	gs.LineWidth = 2.5
	cm := Matrix{A: 5, B: 0, C: 0, D: 5, E: 0, F: 0}
	gs.LastCMMatrix = &cm

	cloned := gs.Clone()

	// Verify all fields are copied
	if cloned.CTM != gs.CTM {
		t.Errorf("Cloned CTM = %+v, want %+v", cloned.CTM, gs.CTM)
	}
	if cloned.ColorSpace != gs.ColorSpace {
		t.Errorf("Cloned ColorSpace = %q, want %q", cloned.ColorSpace, gs.ColorSpace)
	}
	if cloned.StrokeColor != gs.StrokeColor {
		t.Errorf("Cloned StrokeColor = %v, want %v", cloned.StrokeColor, gs.StrokeColor)
	}
	if cloned.FillColor != gs.FillColor {
		t.Errorf("Cloned FillColor = %v, want %v", cloned.FillColor, gs.FillColor)
	}
	if cloned.LineWidth != gs.LineWidth {
		t.Errorf("Cloned LineWidth = %v, want %v", cloned.LineWidth, gs.LineWidth)
	}

	// Modify original; clone should not be affected (for value types this is automatic)
	gs.CTM = Identity()
	gs.ColorSpace = "DeviceGray"
	gs.LineWidth = 99

	if cloned.CTM == gs.CTM {
		t.Error("Cloned CTM should not change when original changes")
	}
	if cloned.ColorSpace == gs.ColorSpace {
		t.Error("Cloned ColorSpace should not change when original changes")
	}
	if cloned.LineWidth == gs.LineWidth {
		t.Error("Cloned LineWidth should not change when original changes")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		expect float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"two values", []float64{3, 7}, 3},
		{"negative values", []float64{-1, -5, 2}, -5},
		{"all same", []float64{4, 4, 4}, 4},
		{"min at end", []float64{10, 20, 5}, 5},
		{"min at start", []float64{1, 10, 20}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := min(tt.values...)
			if got != tt.expect {
				t.Errorf("min(%v) = %v, want %v", tt.values, got, tt.expect)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		expect float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"two values", []float64{3, 7}, 7},
		{"negative values", []float64{-1, -5, 2}, 2},
		{"all same", []float64{4, 4, 4}, 4},
		{"max at start", []float64{20, 10, 5}, 20},
		{"max at end", []float64{1, 10, 20}, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := max(tt.values...)
			if got != tt.expect {
				t.Errorf("max(%v) = %v, want %v", tt.values, got, tt.expect)
			}
		})
	}
}
