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

func IsColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

var colorEnabled = IsColorEnabled()

func Green(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", GreenColor, text, Reset)
}

func Red(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", RedColor, text, Reset)
}

func Yellow(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", YellowColor, text, Reset)
}

func Blue(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", BlueColor, text, Reset)
}

func Cyan(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", CyanColor, text, Reset)
}

func Gray(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", GrayColor, text, Reset)
}

func White(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", WhiteColor, text, Reset)
}

func Purple(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", PurpleColor, text, Reset)
}

func Bold(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", BoldCode, text, Reset)
}

func Success(text string) string {
	return Green("✓ " + text)
}

func Error(text string) string {
	return Red("✗ " + text)
}

func Info(text string) string {
	return Cyan("ℹ " + text)
}

func Warning(text string) string {
	return Yellow("⚠ " + text)
}

func StatusGreen(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", GreenBg, WhiteColor, text, Reset, Reset)
}

func StatusPurple(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", PurpleBg, WhiteColor, text, Reset, Reset)
}

func StatusRed(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", RedBg, WhiteColor, text, Reset, Reset)
}

func StatusYellow(text string) string {
	if !colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", YellowBg, WhiteColor, text, Reset, Reset)
}

