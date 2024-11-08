package logging

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sync"
)

type logLevel int

var (
	once          sync.Once
	infoInstance  *InfoApplication
	errorInstance *ErrorApplication
)

const (
	INFO logLevel = iota
	ERR
)

type skeleton struct {
	level logLevel
}

type ErrorApplication struct {
	skeleton
	logger *log.Logger
}
type InfoApplication struct {
	skeleton
	logger *log.Logger
}

var infoLog = log.New(os.Stdout, "INFO: \t", log.Ltime)
var errorLog = log.New(os.Stderr, "ERROR: \t", log.Ltime)

func newInfoLogger() *InfoApplication {
	once.Do(func() {
		infoInstance = &InfoApplication{
			logger: infoLog,
		}
	})
	return infoInstance
}

func newErrorLogger() *ErrorApplication {
	once.Do(func() {
		errorInstance = &ErrorApplication{
			logger: errorLog,
		}
	})
	return errorInstance
}
func (l *InfoApplication) Info(msg any, args ...any) {

	compiled := assertion(msg)

	if l.level == INFO {
		file, line := getCaller()
		formattedMessage := fmt.Sprintf(compiled, args...) // Format the message using the provided arguments
		l.logger.Printf("\n[%s : %d] \n\n%s\n\n", file, line, formattedMessage)
	}
}

func (l *ErrorApplication) Error(msg any, args ...any) {

	converted := assertion(msg)
	if l.level == ERR {
		file, line := getCaller()
		formattedMessage := fmt.Sprintf(converted, args...) // Format the message using the provided arguments

		l.logger.Fatalf("[%s : %d] \n\n%s\n\n", file, line, formattedMessage)
	}
}

func assertion(msg any) string {
	var comp string

	switch v := msg.(type) {
	case string:
		comp = v
	default:
		comp = fmt.Sprint(v)
	}

	return comp
}

func getCaller() (string, int) {
	key := os.Getenv("KEY_WORD") // assign your folder name, so we can crop no-required part

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		log.Fatal("runtime caller has an error")
	}

	if key == "" {
		fmt.Print("key is empty")
		return file, line // Return without modifying if key is not set
	}

	regExp, _ := regexp.Compile(".*" + regexp.QuoteMeta(key)) // regex for deleting left side

	file = regExp.ReplaceAllString(file, key)

	return file, line
}

var Info = newInfoLogger()
var Err = newErrorLogger()
