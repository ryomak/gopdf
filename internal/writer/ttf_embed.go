package writer

import (
	"bytes"
	"fmt"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/font"
)

// TTFFontEmbedder handles embedding TrueType fonts into PDF
type TTFFontEmbedder struct {
	writer *Writer
}

// NewTTFFontEmbedder creates a new TTF font embedder
func NewTTFFontEmbedder(w *Writer) *TTFFontEmbedder {
	return &TTFFontEmbedder{writer: w}
}

// EmbedTTFFont embeds a TTF font into the PDF and returns a reference to the font object
func (e *TTFFontEmbedder) EmbedTTFFont(ttfFont *font.TTFFont) (*core.Reference, error) {
	// 1. Create FontFile2 stream (embedded TTF data)
	fontFileRef, err := e.createFontFile2(ttfFont)
	if err != nil {
		return nil, fmt.Errorf("failed to create FontFile2: %w", err)
	}

	// 2. Create FontDescriptor
	fontDescriptorRef, err := e.createFontDescriptor(ttfFont, fontFileRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create FontDescriptor: %w", err)
	}

	// 3. Create CIDFont (DescendantFont)
	cidFontRef, err := e.createCIDFont(ttfFont, fontDescriptorRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create CIDFont: %w", err)
	}

	// 4. Create ToUnicode CMap
	toUnicodeRef, err := e.createToUnicodeCMap(ttfFont)
	if err != nil {
		return nil, fmt.Errorf("failed to create ToUnicode CMap: %w", err)
	}

	// 5. Create Type0 font
	type0FontRef, err := e.createType0Font(ttfFont, cidFontRef, toUnicodeRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create Type0 font: %w", err)
	}

	return type0FontRef, nil
}

// createFontFile2 creates a FontFile2 stream object with the TTF data
func (e *TTFFontEmbedder) createFontFile2(ttfFont *font.TTFFont) (*core.Reference, error) {
	fontData := ttfFont.Data()

	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Length"):  core.Integer(len(fontData)),
			core.Name("Length1"): core.Integer(len(fontData)),
		},
		Data: fontData,
	}

	objNum, err := e.writer.AddObject(stream)
	if err != nil {
		return nil, err
	}

	return &core.Reference{
		ObjectNumber:     objNum,
		GenerationNumber: 0,
	}, nil
}

// createFontDescriptor creates a FontDescriptor dictionary
func (e *TTFFontEmbedder) createFontDescriptor(ttfFont *font.TTFFont, fontFileRef *core.Reference) (*core.Reference, error) {
	// Get font metrics from sfnt
	sfntFont := ttfFont.Font()

	// Get font bounding box
	// For simplicity, we use default values
	// In a production implementation, these should be extracted from the font
	fontDescriptor := core.Dictionary{
		core.Name("Type"):        core.Name("FontDescriptor"),
		core.Name("FontName"):    core.Name(ttfFont.Name()),
		core.Name("Flags"):       core.Integer(32), // Symbolic font
		core.Name("FontBBox"):    core.Array{core.Integer(-200), core.Integer(-200), core.Integer(1000), core.Integer(1000)},
		core.Name("ItalicAngle"): core.Integer(0),
		core.Name("Ascent"):      core.Integer(800),
		core.Name("Descent"):     core.Integer(-200),
		core.Name("CapHeight"):   core.Integer(700),
		core.Name("StemV"):       core.Integer(80),
		core.Name("FontFile2"):   fontFileRef,
	}

	// Suppress unused variable warning
	_ = sfntFont

	objNum, err := e.writer.AddObject(fontDescriptor)
	if err != nil {
		return nil, err
	}

	return &core.Reference{
		ObjectNumber:     objNum,
		GenerationNumber: 0,
	}, nil
}

// createCIDSystemInfo creates a CIDSystemInfo dictionary
func (e *TTFFontEmbedder) createCIDSystemInfo() core.Dictionary {
	return core.Dictionary{
		core.Name("Registry"):   core.String("Adobe"),
		core.Name("Ordering"):   core.String("Identity"),
		core.Name("Supplement"): core.Integer(0),
	}
}

// createCIDFont creates a CIDFont (Type 0 descendant font) dictionary
func (e *TTFFontEmbedder) createCIDFont(ttfFont *font.TTFFont, fontDescriptorRef *core.Reference) (*core.Reference, error) {
	cidFont := core.Dictionary{
		core.Name("Type"):           core.Name("Font"),
		core.Name("Subtype"):        core.Name("CIDFontType2"),
		core.Name("BaseFont"):       core.Name(ttfFont.Name()),
		core.Name("CIDSystemInfo"):  e.createCIDSystemInfo(),
		core.Name("FontDescriptor"): fontDescriptorRef,
		// DW (default width) - using 1000 as default
		core.Name("DW"): core.Integer(1000),
		// W array for character widths - simplified for now
		// In production, this should contain actual glyph widths
	}

	objNum, err := e.writer.AddObject(cidFont)
	if err != nil {
		return nil, err
	}

	return &core.Reference{
		ObjectNumber:     objNum,
		GenerationNumber: 0,
	}, nil
}

// createToUnicodeCMap creates a ToUnicode CMap stream
func (e *TTFFontEmbedder) createToUnicodeCMap(ttfFont *font.TTFFont) (*core.Reference, error) {
	// Create a ToUnicode CMap with identity mapping
	// This maps character codes directly to Unicode code points
	var buf bytes.Buffer

	buf.WriteString(`/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<< /Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
`)

	// Add identity mapping for the entire BMP (Basic Multilingual Plane)
	// This maps character code X to Unicode U+X
	buf.WriteString(`100 beginbfrange
`)
	for start := 0x0000; start < 0x10000; start += 0x0100 {
		end := start + 0x00FF
		if end > 0xFFFF {
			end = 0xFFFF
		}
		fmt.Fprintf(&buf, "<%04X> <%04X> <%04X>\n", start, end, start)
	}
	buf.WriteString(`endbfrange
`)

	buf.WriteString(`endcmap
CMapName currentdict /CMap defineresource pop
end
end
`)

	// Suppress unused variable warning
	_ = ttfFont

	cmapData := buf.Bytes()

	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Length"): core.Integer(len(cmapData)),
		},
		Data: cmapData,
	}

	objNum, err := e.writer.AddObject(stream)
	if err != nil {
		return nil, err
	}

	return &core.Reference{
		ObjectNumber:     objNum,
		GenerationNumber: 0,
	}, nil
}

// createType0Font creates a Type0 (composite) font dictionary
func (e *TTFFontEmbedder) createType0Font(ttfFont *font.TTFFont, cidFontRef, toUnicodeRef *core.Reference) (*core.Reference, error) {
	type0Font := core.Dictionary{
		core.Name("Type"):            core.Name("Font"),
		core.Name("Subtype"):         core.Name("Type0"),
		core.Name("BaseFont"):        core.Name(ttfFont.Name()),
		core.Name("Encoding"):        core.Name("Identity-H"),
		core.Name("DescendantFonts"): core.Array{cidFontRef},
		core.Name("ToUnicode"):       toUnicodeRef,
	}

	objNum, err := e.writer.AddObject(type0Font)
	if err != nil {
		return nil, err
	}

	return &core.Reference{
		ObjectNumber:     objNum,
		GenerationNumber: 0,
	}, nil
}
