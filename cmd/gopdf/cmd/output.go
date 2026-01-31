package cmd

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// Icons
const (
	iconSuccess = "✓"
	iconError   = "✗"
	iconWarning = "⚠"
	iconInfo    = "ℹ"
	iconFile    = "📄"
	iconFolder  = "📁"
	iconLock    = "🔒"
	iconUnlock  = "🔓"
	iconKey     = "🔑"
	iconPage    = "📃"
	iconImage   = "🖼"
	iconText    = "📝"
	iconGear    = "⚙"
	iconArrow   = "→"
	iconCheck   = "●"
	iconDot     = "•"
)

// Logger provides styled output
type Logger struct {
	quiet   bool
	verbose bool
	noColor bool
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	return &Logger{
		quiet:   quiet,
		verbose: verbose,
		noColor: os.Getenv("NO_COLOR") != "",
	}
}

func (l *Logger) color(c, s string) string {
	if l.noColor {
		return s
	}
	return c + s + colorReset
}

// Header prints a styled header
func (l *Logger) Header(icon, title string) {
	if l.quiet {
		return
	}
	fmt.Printf("\n%s %s\n", icon, l.color(colorBold, title))
	fmt.Println(l.color(colorDim, strings.Repeat("─", 40)))
}

// Success prints a success message
func (l *Logger) Success(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", l.color(colorGreen, iconSuccess), msg)
}

// Error prints an error message
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s\n", l.color(colorRed, iconError), l.color(colorRed, msg))
}

// Warning prints a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", l.color(colorYellow, iconWarning), l.color(colorYellow, msg))
}

// Info prints an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", l.color(colorCyan, iconInfo), msg)
}

// Detail prints a detail line with indentation
func (l *Logger) Detail(label, value string) {
	if l.quiet {
		return
	}
	fmt.Printf("  %s %s %s\n",
		l.color(colorDim, iconDot),
		l.color(colorDim, label+":"),
		value)
}

// DetailHighlight prints a highlighted detail
func (l *Logger) DetailHighlight(label, value string) {
	if l.quiet {
		return
	}
	fmt.Printf("  %s %s %s\n",
		l.color(colorCyan, iconCheck),
		label+":",
		l.color(colorCyan, value))
}

// Step prints a processing step
func (l *Logger) Step(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", l.color(colorBlue, iconArrow), msg)
}

// Progress prints progress information
func (l *Logger) Progress(current, total int, format string, args ...interface{}) {
	if l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	pct := float64(current) / float64(total) * 100
	fmt.Printf("\r  %s [%3.0f%%] %s", l.color(colorBlue, iconGear), pct, msg)
	if current == total {
		fmt.Println()
	}
}

// FileInfo prints file information
func (l *Logger) FileInfo(filename string) {
	if l.quiet {
		return
	}
	fmt.Printf("%s %s\n", iconFile, l.color(colorBold, filename))
}

// Section prints a section header
func (l *Logger) Section(title string) {
	if l.quiet {
		return
	}
	fmt.Printf("\n%s\n", l.color(colorBold+colorWhite, title))
}

// Table prints a simple table row
func (l *Logger) Table(label string, value interface{}) {
	if l.quiet {
		return
	}
	fmt.Printf("  %-20s %v\n", l.color(colorDim, label), value)
}

// Divider prints a divider line
func (l *Logger) Divider() {
	if l.quiet {
		return
	}
	fmt.Println(l.color(colorDim, strings.Repeat("─", 40)))
}

// Print prints a plain message
func (l *Logger) Print(format string, args ...interface{}) {
	if l.quiet {
		return
	}
	fmt.Printf(format, args...)
}

// Println prints a plain message with newline
func (l *Logger) Println(args ...interface{}) {
	if l.quiet {
		return
	}
	fmt.Println(args...)
}

// Verbose prints a message only in verbose mode
func (l *Logger) Verbose(format string, args ...interface{}) {
	if !l.verbose {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", l.color(colorDim, "[debug]"), l.color(colorDim, msg))
}
