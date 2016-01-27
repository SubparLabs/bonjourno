package log

import (
	"fmt"
)

func Info(msg string, things ...interface{}) {
	fmt.Println(buildLine(msg, "INFO", things...))
}

func Error(msg string, things ...interface{}) {
	fmt.Println(buildLine(msg, "ERROR", things...))
}

func Panic(msg string, things ...interface{}) {
	panic(buildLine(msg, "ERROR", things...))
}

func buildLine(msg string, level string, things ...interface{}) string {
	for i, thing := range things {
		if i%2 == 0 {
			msg = fmt.Sprintf("%s %v=", msg, thing)
		} else {
			msg = fmt.Sprintf("%s%v", msg, thing)
		}
	}

	return fmt.Sprintf("[%5s] %s", level, msg)
}
