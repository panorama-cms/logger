package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const LevelDebug = "DEBUG"
const LevelInfo = "INFO"
const LevelWarning = "WARNING"
const LevelError = "ERROR"
const LevelFatal = "FATAL"

var LogDir = "./logs"
var logDirExists = false
var start = float64(0)
var lastStep = float64(0)

var IncludeRuntime = false
var IncludeStep = false

var LogRequestsSeparately = false
var HideRequestsFromMainLog = false

// init sets some default values by reading the environment variables.
// The following environment variables are supported:
// LOGGER_LOG_DIR: The directory where the log files are stored. Default: ./logs
// LOGGER_INCLUDE_RUNTIME: If set to true, the runtime is included in the log entry. Default: false
// LOGGER_INCLUDE_STEP: If set to true, the step is included in the log entry. Default: false
// LOGGER_LOG_REQUESTS_SEPARATELY: If set to true, the requests are logged in a separate file. Default: false
// LOGGER_HIDE_REQUESTS_FROM_MAIN_LOG: If set to true, the requests are not logged in the main log file. Default: false
func init() {
	logDirTemp := os.Getenv("LOGGER_LOG_DIR")
	logDirTemp = strings.TrimSpace(logDirTemp)
	if logDirTemp != "" {
		LogDir = logDirTemp
	}

	includeRuntimeTemp := os.Getenv("LOGGER_INCLUDE_RUNTIME")
	includeRuntimeTemp = strings.TrimSpace(includeRuntimeTemp)
	if includeRuntimeTemp == "true" {
		IncludeRuntime = true
	}

	includeStepTemp := os.Getenv("LOGGER_INCLUDE_STEP")
	includeStepTemp = strings.TrimSpace(includeStepTemp)
	if includeStepTemp == "true" {
		IncludeStep = true
	}

	logRequestsSeparatelyTemp := os.Getenv("LOGGER_LOG_REQUESTS_SEPARATELY")
	logRequestsSeparatelyTemp = strings.TrimSpace(logRequestsSeparatelyTemp)
	if logRequestsSeparatelyTemp == "true" {
		LogRequestsSeparately = true
	}

	hideRequestsFromMainLogTemp := os.Getenv("LOGGER_HIDE_REQUESTS_FROM_MAIN_LOG")
	hideRequestsFromMainLogTemp = strings.TrimSpace(hideRequestsFromMainLogTemp)
	if hideRequestsFromMainLogTemp == "true" {
		HideRequestsFromMainLog = true
	}

	// check if logs directory exists, if not create it
	_, err := os.Stat(LogDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(LogDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		logDirExists = true
	}
}

// microTime returns the current time in microseconds.
func microTime() float64 {
	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)
	micSeconds := float64(now.Nanosecond()) / 1000000000
	return float64(now.Unix()) + micSeconds
}

// formatMicroTimeDuration formats a duration in microseconds to a string.
// The format is DD:HH:MM:SS.MICROSECONDS
func formatMicroTimeDuration(duration float64) string {
	// Format: DD:HH:MM:SS.MICROSECONDS
	formatString := "%02d:%02d:%02d:%02d.%06d"

	days := int(duration / 86400)
	duration -= float64(days * 86400)

	hours := int(duration / 3600)
	duration -= float64(hours * 3600)

	minutes := int(duration / 60)
	duration -= float64(minutes * 60)

	seconds := int(duration)
	duration -= float64(seconds)

	microSeconds := int(duration * 1000000)
	return fmt.Sprintf(formatString, days, hours, minutes, seconds, microSeconds)
}

// l is the main logging function.
// It logs the given content to the main log file.
// It's internal and should not be used directly because we provide wrapper functions for each log level below.
func l(level string, content string) {
	if level == "" {
		level = LevelInfo
	} else if level != LevelDebug && level != LevelInfo && level != LevelWarning && level != LevelError && level != LevelFatal {
		level = LevelInfo
	}

	if !logDirExists {
		// check if directory logs exists, if not create it
		_, err := os.Stat(LogDir)
		if os.IsNotExist(err) {
			err = os.Mkdir(LogDir, 0755)
			if err != nil {
				log.Fatal(err)
			}
			logDirExists = true
		}
	}

	// get the current date
	t := time.Now()

	// format time to YYYY-MM-DD
	date := t.Format("2006-01-02")

	// format time to HH:MM:SS
	tFormatted := t.Format("2006-01-02 15:04:05.000000")

	// open file YYYY-MM-DD.log
	f, err := os.OpenFile(LogDir+"/"+date+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if start == 0 {
		start = microTime()
		lastStep = start
	}

	runtime := microTime() - start
	step := microTime() - lastStep
	lastStep = microTime()

	var runtimeFormatted string
	var stepFormatted string

	runtimeFormatted = formatMicroTimeDuration(runtime)
	stepFormatted = formatMicroTimeDuration(step)

	entry := "[" + tFormatted + "]"
	if IncludeRuntime {
		entry += "[" + runtimeFormatted + "]"
	}
	if IncludeStep {
		entry += "[" + stepFormatted + "]"
	}

	entry += " " + level + " " + content + "\n"

	// write to file
	_, err = f.WriteString(entry)
	if err != nil {
		log.Fatal(err)
	}

	// close file
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// Log logs a message with the given log level.
func Log(level string, content string) {
	l(level, content)
}

// LogAsync logs a message with the given log level asynchronously by calling logger.l as goroutine.
func LogAsync(level string, content string) {
	go l(level, content)
}

// Debug logs a debug message.
func Debug(content string) {
	l(LevelDebug, content)
}

// DebugAsync logs a debug message asynchronously by calling logger.l as goroutine.
func DebugAsync(content string) {
	go l(LevelDebug, content)
}

// Info logs an info message.
func Info(content string) {
	l(LevelInfo, content)
}

// InfoAsync logs an info message asynchronously by calling logger.l as goroutine.
func InfoAsync(content string) {
	go l(LevelInfo, content)
}

// Warning logs a warning message.
func Warning(content string) {
	l(LevelWarning, content)
}

// WarningAsync logs a warning message asynchronously by calling logger.l as goroutine.
func WarningAsync(content string) {
	go l(LevelWarning, content)
}

// Error logs an error message.
func Error(content string) {
	l(LevelError, content)
}

// ErrorAsync logs an error message asynchronously by calling logger.l as goroutine.
func ErrorAsync(content string) {
	go l(LevelError, content)
}

// Fatal logs a fatal message.
func Fatal(content string) {
	l(LevelFatal, content)
}

// FatalAsync logs a fatal message asynchronously by calling logger.l as goroutine.
func FatalAsync(content string) {
	go l(LevelFatal, content)
}

// LogRequest logs a request.
// If HideRequestsFromMainLog is true, the request will not be logged to the main log file but only when LogRequestsSeparately is true.
func LogRequest(method string, path string, userAgent string, ip string) {
	if (!LogRequestsSeparately) || (LogRequestsSeparately && !HideRequestsFromMainLog) {
		Log(LevelInfo, fmt.Sprintf("(%s) %s <- %s @ %s", method, path, userAgent, ip))
	}

	if LogRequestsSeparately {
		// get the current date
		t := time.Now()

		// format time to YYYY-MM-DD
		date := t.Format("2006-01-02")

		// format time to HH:MM:SS
		tFormatted := t.Format("2006-01-02 15:04:05.000000")

		// open file requests.csv
		f, err := os.OpenFile(LogDir+"/requests-"+date+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// replace all , with ; in user agent
		userAgent = strings.ReplaceAll(userAgent, ",", ";")

		entry := tFormatted + "," + method + "," + path + "," + userAgent + "," + ip + "\n"

		// write to file
		_, err = f.WriteString(entry)
		if err != nil {
			log.Fatal(err)
		}
	}
}
