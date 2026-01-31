package font

// StandardFontMap maps PDF font names to StandardFont values.
var StandardFontMap = map[string]StandardFont{
	// Helvetica variants
	"Helvetica":             Helvetica,
	"Helvetica-Bold":        HelveticaBold,
	"Helvetica-Oblique":     HelveticaOblique,
	"Helvetica-BoldOblique": HelveticaBoldOblique,
	// Times variants
	"Times-Roman":      TimesRoman,
	"Times-Bold":       TimesBold,
	"Times-Italic":     TimesItalic,
	"Times-BoldItalic": TimesBoldItalic,
	// Courier variants
	"Courier":             Courier,
	"Courier-Bold":        CourierBold,
	"Courier-Oblique":     CourierOblique,
	"Courier-BoldOblique": CourierBoldOblique,
	// Symbol fonts
	"Symbol":       Symbol,
	"ZapfDingbats": ZapfDingbats,
}

// MapToStandardFont maps a PDF font name to a StandardFont.
// Returns the mapped font and true if found, otherwise empty string and false.
func MapToStandardFont(fontName string, isBold bool) (StandardFont, bool) {
	// Direct match
	if stdFont, ok := StandardFontMap[fontName]; ok {
		return stdFont, true
	}

	// Partial match to guess font family
	// Example: "BCDEEE+Helvetica-Bold" -> "Helvetica-Bold"
	for name, stdFont := range StandardFontMap {
		if len(fontName) >= len(name) && fontName[len(fontName)-len(name):] == name {
			return stdFont, true
		}
	}

	// Guess family from font name
	switch {
	case ContainsFont(fontName, "Helvetica"):
		if isBold {
			return HelveticaBold, true
		}
		return Helvetica, true
	case ContainsFont(fontName, "Times"):
		if isBold {
			return TimesBold, true
		}
		return TimesRoman, true
	case ContainsFont(fontName, "Courier"):
		if isBold {
			return CourierBold, true
		}
		return Courier, true
	case ContainsFont(fontName, "Symbol"):
		return Symbol, true
	case ContainsFont(fontName, "ZapfDingbats"), ContainsFont(fontName, "Dingbats"):
		return ZapfDingbats, true
	}

	return "", false
}

// ContainsFont checks if fontName contains the target string (case-insensitive).
func ContainsFont(fontName, target string) bool {
	// Case-insensitive partial match
	for i := 0; i <= len(fontName)-len(target); i++ {
		match := true
		for j := 0; j < len(target); j++ {
			c1 := fontName[i+j]
			c2 := target[j]
			// Case-insensitive comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
