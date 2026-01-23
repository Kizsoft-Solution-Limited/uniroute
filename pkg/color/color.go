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
	// Background colors
	GreenBg   = "\033[42m"
	PurpleBg  = "\033[45m"
	RedBg     = "\033[41m"
	YellowBg  = "\033[43m"
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

// White returns white colored text
func White(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", WhiteColor, text, Reset)
}

// Purple returns purple colored text
func Purple(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", PurpleColor, text, Reset)
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

// StatusGreen returns white text on green background (for 200 OK status)
func StatusGreen(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", GreenBg, WhiteColor, text, Reset, Reset)
}

// StatusPurple returns white text on purple background (for 201 Created status)
func StatusPurple(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", PurpleBg, WhiteColor, text, Reset, Reset)
}

// StatusRed returns white text on red background (for error status)
func StatusRed(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", RedBg, WhiteColor, text, Reset, Reset)
}

// StatusYellow returns white text on yellow background (for warning status)
func StatusYellow(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", YellowBg, WhiteColor, text, Reset, Reset)
}

