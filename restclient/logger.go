package restclient

import (
	"fmt"
	"log"
	"os"
)

const (
	red color = iota + 31
	green
	yellow
	blue
)

// color represents a text color.
type color uint8

// Add adds the coloring to the given string.
func (c color) add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

/* infoLogger, debugLogger, warningLogger and errorLogger are internal private loggers to log requests
3rd party go logging libraries have been avoided intentionally to omit unnecessary dependencies on the user. */
var (
	infoLogger    *log.Logger
	debugLogger   *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func init() {
	infoLogger = log.New(os.Stdout, blue.add(" [ INFO ] "), log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(os.Stdout, green.add(" [ DEBUG ] "), log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger = log.New(os.Stdout, yellow.add(" [ WARN ] "), log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stdout, red.add(" [ ERROR ] "), log.Ldate|log.Ltime|log.Lshortfile)
}
