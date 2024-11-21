package logging

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sync"
)

// variables for singletons and instances
var (
	infoOnce      sync.Once
	errorOnce     sync.Once
	infoInstance  *InfoApplication
	errorInstance *ErrorApplication
)

// types
type InfoApplication struct {
	log *log.Logger
}

type ErrorApplication struct {
	log *log.Logger
}

// NOTE base logger instances
var infoLog = log.New(os.Stdin, "INFO \t", log.Ltime)
var errorLog = log.New(os.Stderr, "ERROR \t", log.Ltime)

// NOTE info and error singletons
func InfoSingleTon() *InfoApplication {
	infoOnce.Do(func() {
		infoInstance = &InfoApplication{
			log: infoLog,
		}
	})

	return infoInstance
}

func ErrorSingleTon() *ErrorApplication {
	errorOnce.Do(func() {
		errorInstance = &ErrorApplication{
			log: errorLog,
		}
	})

	return errorInstance
}

func (l *InfoApplication) Info(msg any, args ...any) {
	// assertion
	asserted := fmt.Sprint(msg)
	file, line := getCaller()
	formatted := fmt.Sprintf(asserted, args...)
	l.log.Printf("[%s : %d] \n\n%s\n\n", file, line, formatted)
}

func (l *ErrorApplication) Error(msg any, args ...any) {
	asserted := fmt.Sprint(msg)
	file, line := getCaller()
	formatted := fmt.Sprintf(asserted, args...)
	l.log.Panicf("[%s : %d] \n\n%s\n\n", file, line, formatted)
}

func getCaller() (string, int) {

	projectRootFolder := os.Getenv("KEY_WORD")

	_, file, line, ok := runtime.Caller(4)
	if !ok {
		log.Fatalln("caller can not be defined")
	}

	r, _ := regexp.Compile(".*" + regexp.QuoteMeta(projectRootFolder))
	filterFileString := r.ReplaceAllString(file, projectRootFolder)
	return filterFileString, line
}

var (
	Info  = InfoSingleTon()
	Error = ErrorSingleTon()
)
