package log

import (
	"fmt"
)

func Info(msg string, things ...interface{}) {
	log(msg, "INFO", things...)
}

func Error(msg string, things ...interface{}) {
	log(msg, "ERROR", things...)
}

func log(msg string, level string, things ...interface{}) {
	for i, thing := range things {
		if i%2 == 0 {
			msg = fmt.Sprintf("%s %v=", msg, thing)
		} else {
			msg = fmt.Sprintf("%s%v", msg, thing)
		}
	}

	fmt.Printf("[%5s] %s\n", level, msg)
}
