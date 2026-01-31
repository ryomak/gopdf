package content

import (
	"testing"
)

func TestToUnicodeCMap_Lookup_CharMap(t *testing.T) {
	cmap := &ToUnicodeCMap{
		charMap: map[uint16]rune{
			0x0026: 0x0048, // CID 38 -> 'H'
			0x004f: 0x0065, // CID 79 -> 'e'
			0x0048: 0x006c, // CID 72 -> 'l'
		},
	}

	tests := []struct {
		name     string
		cid      uint16
		wantChar rune
		wantOk   bool
	}{
		{"Found in charMap", 0x0026, 0x0048, true},
		{"Another in charMap", 0x004f, 0x0065, true},
		{"Not found", 0x9999, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChar, gotOk := cmap.Lookup(tt.cid)
			if gotChar != tt.wantChar || gotOk != tt.wantOk {
				t.Errorf("Lookup(%04X) = (%04X, %v), want (%04X, %v)",
					tt.cid, gotChar, gotOk, tt.wantChar, tt.wantOk)
			}
		})
	}
}

func TestToUnicodeCMap_Lookup_Ranges(t *testing.T) {
	cmap := &ToUnicodeCMap{
		charMap: map[uint16]rune{},
		ranges: []CIDRange{
			{StartCID: 0x1000, EndCID: 0x1005, StartChar: 0x4e00}, // 范围映射
		},
	}

	tests := []struct {
		name     string
		cid      uint16
		wantChar rune
		wantOk   bool
	}{
		{"Start of range", 0x1000, 0x4e00, true},
		{"Middle of range", 0x1003, 0x4e03, true},
		{"End of range", 0x1005, 0x4e05, true},
		{"Before range", 0x0fff, 0, false},
		{"After range", 0x1006, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChar, gotOk := cmap.Lookup(tt.cid)
			if gotChar != tt.wantChar || gotOk != tt.wantOk {
				t.Errorf("Lookup(%04X) = (%04X, %v), want (%04X, %v)",
					tt.cid, gotChar, gotOk, tt.wantChar, tt.wantOk)
			}
		})
	}
}

func TestToUnicodeCMap_Lookup_Nil(t *testing.T) {
	var cmap *ToUnicodeCMap
	gotChar, gotOk := cmap.Lookup(0x0001)
	if gotChar != 0 || gotOk != false {
		t.Errorf("Lookup on nil CMap should return (0, false), got (%04X, %v)", gotChar, gotOk)
	}
}

func TestToUnicodeCMap_LookupString(t *testing.T) {
	cmap := &ToUnicodeCMap{
		charMap: map[uint16]rune{
			0x0026: 0x0048, // 'H'
			0x004f: 0x0065, // 'e'
			0x0048: 0x006c, // 'l'
			0x0050: 0x006f, // 'o'
		},
	}

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "Hello",
			data:     []byte{0x00, 0x26, 0x00, 0x4f, 0x00, 0x48, 0x00, 0x48, 0x00, 0x50}, // CIDs for "Hello"
			expected: "Hello",
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: "",
		},
		{
			name:     "Single character",
			data:     []byte{0x00, 0x26}, // 'H'
			expected: "H",
		},
		{
			name:     "Odd length (incomplete)",
			data:     []byte{0x00, 0x26, 0x00}, // Should skip incomplete byte
			expected: "H",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmap.LookupString(tt.data)
			if result != tt.expected {
				t.Errorf("LookupString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestToUnicodeCMap_LookupString_Nil(t *testing.T) {
	var cmap *ToUnicodeCMap
	result := cmap.LookupString([]byte{0x00, 0x01})
	if result != "" {
		t.Errorf("LookupString on nil CMap should return empty string, got %q", result)
	}
}

func TestParseBFChar(t *testing.T) {
	// Simulate a simple bfchar section
	data := []byte(`
100 beginbfchar
<0026> <0048>
<004f> <0065>
<0048> <006c>
endbfchar
`)

	result, err := parseBFChar(data)
	if err != nil {
		t.Fatalf("parseBFChar failed: %v", err)
	}

	expected := map[uint16]rune{
		0x0026: 0x0048,
		0x004f: 0x0065,
		0x0048: 0x006c,
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d mappings, got %d", len(expected), len(result))
	}

	for cid, unicode := range expected {
		if result[cid] != unicode {
			t.Errorf("Expected CID %04X -> %04X, got %04X", cid, unicode, result[cid])
		}
	}
}

func TestParseBFChar_MultipleSections(t *testing.T) {
	data := []byte(`
10 beginbfchar
<0001> <0041>
<0002> <0042>
endbfchar
5 beginbfchar
<0003> <0043>
endbfchar
`)

	result, err := parseBFChar(data)
	if err != nil {
		t.Fatalf("parseBFChar failed: %v", err)
	}

	expected := map[uint16]rune{
		0x0001: 0x0041,
		0x0002: 0x0042,
		0x0003: 0x0043,
	}

	for cid, unicode := range expected {
		if result[cid] != unicode {
			t.Errorf("Expected CID %04X -> %04X, got %04X", cid, unicode, result[cid])
		}
	}
}

func TestParseBFRange(t *testing.T) {
	data := []byte(`
10 beginbfrange
<1000> <1005> <4e00>
<2000> <2003> <5000>
endbfrange
`)

	result, err := parseBFRange(data)
	if err != nil {
		t.Fatalf("parseBFRange failed: %v", err)
	}

	expected := []CIDRange{
		{StartCID: 0x1000, EndCID: 0x1005, StartChar: 0x4e00},
		{StartCID: 0x2000, EndCID: 0x2003, StartChar: 0x5000},
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d ranges, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		got := result[i]
		if got.StartCID != exp.StartCID || got.EndCID != exp.EndCID || got.StartChar != exp.StartChar {
			t.Errorf("Range[%d]: got {%04X, %04X, %04X}, want {%04X, %04X, %04X}",
				i, got.StartCID, got.EndCID, got.StartChar,
				exp.StartCID, exp.EndCID, exp.StartChar)
		}
	}
}

func TestParseToUnicodeCMap(t *testing.T) {
	// Full CMap example
	data := []byte(`
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo <<
  /Registry (Adobe)
  /Ordering (UCS)
  /Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <ffff>
endcodespacerange

10 beginbfchar
<0026> <0048>
<004f> <0065>
<0048> <006c>
endbfchar

5 beginbfrange
<1000> <1005> <4e00>
endbfrange

endcmap
CMapName currentdict /CMap defineresource pop
end
end
`)

	cmap, err := ParseToUnicodeCMap(data)
	if err != nil {
		t.Fatalf("ParseToUnicodeCMap failed: %v", err)
	}

	// Test charMap
	if len(cmap.charMap) != 3 {
		t.Errorf("Expected 3 char mappings, got %d", len(cmap.charMap))
	}

	if cmap.charMap[0x0026] != 0x0048 {
		t.Errorf("Expected CID 0x0026 -> 0x0048, got %04X", cmap.charMap[0x0026])
	}

	// Test ranges
	if len(cmap.ranges) != 1 {
		t.Errorf("Expected 1 range, got %d", len(cmap.ranges))
	}

	if cmap.ranges[0].StartCID != 0x1000 {
		t.Errorf("Expected range StartCID 0x1000, got %04X", cmap.ranges[0].StartCID)
	}

	// Test lookup
	char, ok := cmap.Lookup(0x0026)
	if !ok || char != 0x0048 {
		t.Errorf("Lookup(0x0026) failed: got (%04X, %v)", char, ok)
	}

	char, ok = cmap.Lookup(0x1003)
	if !ok || char != 0x4e03 {
		t.Errorf("Lookup(0x1003) in range failed: got (%04X, %v)", char, ok)
	}
}

func TestExtractHexPairs(t *testing.T) {
	data := []byte("<0026> <0048> <004f> <0065>")
	result := extractHexPairs(data)

	expected := [][]string{
		{"0026", "0048"},
		{"004f", "0065"},
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d pairs, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		got := result[i]
		if len(got) != 2 || got[0] != exp[0] || got[1] != exp[1] {
			t.Errorf("Pair[%d]: got %v, want %v", i, got, exp)
		}
	}
}

func TestExtractHexTriples(t *testing.T) {
	data := []byte("<1000> <1005> <4e00> <2000> <2003> <5000>")
	result := extractHexTriples(data)

	expected := [][]string{
		{"1000", "1005", "4e00"},
		{"2000", "2003", "5000"},
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d triples, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		got := result[i]
		if len(got) != 3 || got[0] != exp[0] || got[1] != exp[1] || got[2] != exp[2] {
			t.Errorf("Triple[%d]: got %v, want %v", i, got, exp)
		}
	}
}

func TestParseHex(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0026", 38},
		{"004f", 79},
		{"FFFF", 65535},
		{"4e00", 19968},
		{"invalid", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseHex(tt.input)
			if result != tt.expected {
				t.Errorf("parseHex(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
