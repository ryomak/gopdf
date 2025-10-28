package content

// Matrix は変換行列（CTM: Current Transformation Matrix）
type Matrix struct {
	A, B, C, D, E, F float64 // [a b c d e f]
}

// Identity は単位行列を返す
func Identity() Matrix {
	return Matrix{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}
}

// Multiply は行列の乗算を行う
// 新しいCTM = 現在のCTM × 新しい変換行列
func (m Matrix) Multiply(other Matrix) Matrix {
	return Matrix{
		A: m.A*other.A + m.B*other.C,
		B: m.A*other.B + m.B*other.D,
		C: m.C*other.A + m.D*other.C,
		D: m.C*other.B + m.D*other.D,
		E: m.E*other.A + m.F*other.C + other.E,
		F: m.E*other.B + m.F*other.D + other.F,
	}
}

// TransformPoint は座標を変換する
func (m Matrix) TransformPoint(x, y float64) (float64, float64) {
	return m.A*x + m.C*y + m.E, m.B*x + m.D*y + m.F
}

// TransformRect は矩形を変換する（4隅を変換）
func (m Matrix) TransformRect(x, y, width, height float64) (minX, minY, maxX, maxY float64) {
	// 4隅の座標
	x1, y1 := m.TransformPoint(x, y)
	x2, y2 := m.TransformPoint(x+width, y)
	x3, y3 := m.TransformPoint(x, y+height)
	x4, y4 := m.TransformPoint(x+width, y+height)

	// 最小・最大を求める
	minX = min(x1, x2, x3, x4)
	maxX = max(x1, x2, x3, x4)
	minY = min(y1, y2, y3, y4)
	maxY = max(y1, y2, y3, y4)

	return
}

// GraphicsState は現在のグラフィックス状態
type GraphicsState struct {
	CTM         Matrix      // Current Transformation Matrix
	ColorSpace  string      // 色空間
	StrokeColor [3]float64  // 線の色（RGB）
	FillColor   [3]float64  // 塗りつぶし色（RGB）
	LineWidth   float64     // 線幅
}

// NewGraphicsState は新しいGraphicsStateを作成する
func NewGraphicsState() GraphicsState {
	return GraphicsState{
		CTM:         Identity(),
		ColorSpace:  "DeviceRGB",
		StrokeColor: [3]float64{0, 0, 0},
		FillColor:   [3]float64{0, 0, 0},
		LineWidth:   1.0,
	}
}

// Clone はGraphicsStateのコピーを作成する（スタック用）
func (gs GraphicsState) Clone() GraphicsState {
	return gs
}

func min(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	result := values[0]
	for _, v := range values[1:] {
		if v < result {
			result = v
		}
	}
	return result
}

func max(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	result := values[0]
	for _, v := range values[1:] {
		if v > result {
			result = v
		}
	}
	return result
}
