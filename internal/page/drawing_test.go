package page

import (
	"bytes"
	"strings"
	"testing"
)

func TestDrawTextInternal(t *testing.T) {
	tests := []struct {
		name        string
		x, y        float64
		fontKey     string
		fontSize    float64
		encodedText string
		useBrackets bool
		wantOps     []string
	}{
		{
			name:        "standard font with brackets",
			x:           100.0,
			y:           200.0,
			fontKey:     "F1",
			fontSize:    12.0,
			encodedText: "Hello",
			useBrackets: true,
			wantOps:     []string{"BT", "/F1 12.00 Tf", "100.00 200.00 Td", "(Hello) Tj", "ET"},
		},
		{
			name:        "TTF font with hex encoding",
			x:           50.0,
			y:           100.0,
			fontKey:     "F15",
			fontSize:    14.0,
			encodedText: "0041004200",
			useBrackets: false,
			wantOps:     []string{"BT", "/F15 14.00 Tf", "<0041004200> Tj", "ET"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			DrawTextInternal(&buf, tt.x, tt.y, tt.fontKey, tt.fontSize, tt.encodedText, tt.useBrackets)

			result := buf.String()
			for _, op := range tt.wantOps {
				if !strings.Contains(result, op) {
					t.Errorf("expected operator %q in output, got:\n%s", op, result)
				}
			}
		})
	}
}

func TestDrawCirclePath(t *testing.T) {
	tests := []struct {
		name          string
		centerX       float64
		centerY       float64
		radius        float64
		wantMoveTo    bool
		wantCurveOps  int
	}{
		{
			name:         "unit circle at origin",
			centerX:      0,
			centerY:      0,
			radius:       1.0,
			wantMoveTo:   true,
			wantCurveOps: 4,
		},
		{
			name:         "circle with offset",
			centerX:      100,
			centerY:      200,
			radius:       50.0,
			wantMoveTo:   true,
			wantCurveOps: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			DrawCirclePath(&buf, tt.centerX, tt.centerY, tt.radius)

			result := buf.String()

			// Check for move to operator
			if tt.wantMoveTo && !strings.Contains(result, " m\n") {
				t.Error("expected move to operator 'm' in output")
			}

			// Count curve operators
			curveCount := strings.Count(result, " c\n")
			if curveCount != tt.wantCurveOps {
				t.Errorf("expected %d curve operators, got %d", tt.wantCurveOps, curveCount)
			}
		})
	}
}
