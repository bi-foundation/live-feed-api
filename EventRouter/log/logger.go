package log

import (
	"log"
	"os"
	"strings"
)

// Level of the logger
type Level int8

const (
	// D debug
	D Level = iota
	// I info
	I
	// W warning
	W
	// E error
	E
	// F fail
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

// Parse a string to a Level. Return D as default
func Parse(lvl string) Level {
	return toLevel[strings.ToLower(lvl)]
}

// SetLevel of the logger
func SetLevel(lvl Level) {
	level = lvl
}

// SetLogger to use a custom logger
func SetLogger(l *log.Logger) {
	logger = l
}

// Debug log on debug level
func Debug(format string, v ...interface{}) {
	if level <= D {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info log on info level
func Info(format string, v ...interface{}) {
	if level <= I {
		logger.Printf("[INFO] "+format, v...)
	}
}

// Warn log on warning level
func Warn(format string, v ...interface{}) {
	if level <= W {
		logger.Printf("[WARN] "+format, v...)
	}
}

// Error log on error level
func Error(format string, v ...interface{}) {
	if level <= E {
		logger.Printf("[ERROR] "+format, v...)
	}
}

// Fatal log on fatal level
func Fatal(format string, v ...interface{}) {
	if level <= F {
		logger.Fatalf("[FATAL] "+format, v...)
	}
}
