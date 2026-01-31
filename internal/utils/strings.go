package utils

// ReplaceAll replaces all occurrences of old with new in string s.
// This is a pure Go implementation without using strings package.
func ReplaceAll(s, old, new string) string {
	if old == "" {
		return s
	}

	result := ""
	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}
