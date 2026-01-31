package utils

import (
	"testing"
)

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		name string
		s    string
		old  string
		new  string
		want string
	}{
		{
			name: "simple replace",
			s:    "hello world",
			old:  "world",
			new:  "gopher",
			want: "hello gopher",
		},
		{
			name: "multiple replacements",
			s:    "a-b-c-d",
			old:  "-",
			new:  "_",
			want: "a_b_c_d",
		},
		{
			name: "no match",
			s:    "hello",
			old:  "x",
			new:  "y",
			want: "hello",
		},
		{
			name: "empty string",
			s:    "",
			old:  "a",
			new:  "b",
			want: "",
		},
		{
			name: "empty old (no replace)",
			s:    "hello",
			old:  "",
			new:  "x",
			want: "hello",
		},
		{
			name: "replace with empty",
			s:    "a-b-c",
			old:  "-",
			new:  "",
			want: "abc",
		},
		{
			name: "overlapping pattern",
			s:    "aaaa",
			old:  "aa",
			new:  "b",
			want: "bb",
		},
		{
			name: "escape backslash",
			s:    "path\\file",
			old:  "\\",
			new:  "\\\\",
			want: "path\\\\file",
		},
		{
			name: "escape parenthesis",
			s:    "(test)",
			old:  "(",
			new:  "\\(",
			want: "\\(test)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceAll(tt.s, tt.old, tt.new)
			if got != tt.want {
				t.Errorf("ReplaceAll(%q, %q, %q) = %q, want %q", tt.s, tt.old, tt.new, got, tt.want)
			}
		})
	}
}
