package log

import (
	"log"
	"os"
	"strings"
)

type Level int8

const (
	D Level = iota
	I
	W
	E
	F
)

var toLevel = map[string]Level{
	"d":       D,
	"debug":   D,
	"i":       I,
	"info":    I,
	"w":       W,
	"warning": W,
	"e":       E,
	"error":   E,
	"f":       F,
	"fatal":   F,
}

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime)
var level = I

func Parse(lvl string) Level {
	return toLevel[strings.ToLower(lvl)]
}

func SetLevel(lvl Level) {
	level = lvl
}
func SetLogger(l *log.Logger) {
	logger = l
}

func Debug(format string, v ...interface{}) {
	if level <= D {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if level <= I {
		logger.Printf("[INFO] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if level <= W {
		logger.Printf("[WARN] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if level <= E {
		logger.Printf("[ERROR] "+format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	if level <= F {
		logger.Fatalf("[FATAL] "+format, v...)
	}
}
