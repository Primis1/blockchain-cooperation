package logging

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
)

type logLevel int

const (
	INFO logLevel = iota
	ERR
)

type Application struct {
	Level    logLevel
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

type LoggingFacadeT struct {
	Log *Application
}

var InfoLog = log.New(os.Stdout, "INFO: \t", log.Ltime)
var ErrorLog = log.New(os.Stderr, "ERROR: \t", log.Ltime)

var path string

var LoggingFacade = &LoggingFacadeT{}

func (l *LoggingFacadeT) NewLogger(Level logLevel) *Application {
	return &Application{
		Level:    Level,
		ErrorLog: ErrorLog,
		InfoLog:  InfoLog,
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

func (l *LoggingFacadeT) Info(msg any, args ...any) {

	compiled := assertion(msg)

	if l.Log.Level == INFO {
		file, line := getCaller()
		formattedMessage := fmt.Sprintf(compiled, args...) // Format the message using the provided arguments
		l.Log.InfoLog.Printf("[%s : %d] \n\n%s\n\n", file, line, formattedMessage)
	}
}

func (l *LoggingFacadeT) Error(msg any, args ...any) {

	converted := assertion(msg)
	if l.Log.Level == ERR {
		file, line := getCaller()
		formattedMessage := fmt.Sprintf(converted, args...) // Format the message using the provided arguments

		l.Log.ErrorLog.Fatalf("[%s : %d] \n\n%s\n\n", file, line, formattedMessage)
	}
}

func getCaller() (string, int) {
	key := os.Getenv("KEY_WORD") // assign your folder name, so we can crop no-reqired part

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
