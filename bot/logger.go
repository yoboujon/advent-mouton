package bot

import (
	"fmt"
	"time"
)

type colorFormat string
type LogLevel int

const (
	BLUE      colorFormat = "\033[34m"
	GREEN     colorFormat = "\033[32m"
	YELLOW    colorFormat = "\033[33m"
	RED       colorFormat = "\033[31m"
	LIGHT_RED colorFormat = "\033[91m"
	WHITE     colorFormat = "\033[97m"
	RESET     colorFormat = "\033[0m"
)

const (
	NOLOG LogLevel = iota
	INFO
	CORRECT
	WARNING
	ERROR
)

var colorPrint = map[LogLevel]colorFormat{
	NOLOG:   "",
	INFO:    BLUE + "[INFO]" + RESET,
	CORRECT: GREEN,
	WARNING: YELLOW + "[WARN]",
	ERROR:   RED + "[ERR.]" + LIGHT_RED,
}

var minimumlog = NOLOG

func Logformat(level LogLevel, format string, args ...interface{}) {
	if level < minimumlog {
		return
	}

	header := colorPrint[level]
	message := fmt.Sprintf(format, args...)
	currentTime := time.Now()
	time := fmt.Sprintf("[%02d/%02d/%d %02d:%02d:%02d]", currentTime.Day(), currentTime.Month(), currentTime.Year(), currentTime.Hour(), currentTime.Minute(), currentTime.Second())
	if level == NOLOG {
		fmt.Printf("%s\t\t%s", time, message)
	} else if level != INFO {
		fmt.Printf("%s%s\t%s%s", header, time, message, RESET)
	} else {
		fmt.Printf("%s%s\t%s", header, time, message)
	}
}

func SetLevel(level LogLevel) {
	minimumlog = level
}
