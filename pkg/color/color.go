package color

import (
	"fmt"
	"os"
	"runtime"
)

// ANSI color codes
const (
	Reset      = "\033[0m"
	RedColor   = "\033[31m"
	GreenColor = "\033[32m"
	YellowColor = "\033[33m"
	BlueColor  = "\033[34m"
	PurpleColor = "\033[35m"
	CyanColor  = "\033[36m"
	WhiteColor = "\033[37m"
	GrayColor  = "\033[90m"
	BoldCode   = "\033[1m"
)

// IsColorEnabled checks if colors should be enabled
func IsColorEnabled() bool {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	// Check if output is a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	
	// On Windows, colors need special handling
	if runtime.GOOS == "windows" {
		// For now, disable on Windows (can be enhanced with Windows API)
		return false
	}
	
	// Check if it's a character device (terminal)
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

var colorEnabled = IsColorEnabled()

// Green returns green colored text
func Green(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", GreenColor, text, Reset)
}

// Red returns red colored text
func Red(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", RedColor, text, Reset)
}

// Yellow returns yellow colored text
func Yellow(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", YellowColor, text, Reset)
}

// Blue returns blue colored text
func Blue(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", BlueColor, text, Reset)
}

// Cyan returns cyan colored text
func Cyan(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", CyanColor, text, Reset)
}

// Gray returns gray colored text
func Gray(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", GrayColor, text, Reset)
}

// Bold returns bold text
func Bold(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", BoldCode, text, Reset)
}

// Success returns green text with checkmark
func Success(text string) string {
	return Green("✓ " + text)
}

// Error returns red text with X
func Error(text string) string {
	return Red("✗ " + text)
}

// Info returns cyan text with info icon
func Info(text string) string {
	return Cyan("ℹ " + text)
}

// Warning returns yellow text with warning icon
func Warning(text string) string {
	return Yellow("⚠ " + text)
}

