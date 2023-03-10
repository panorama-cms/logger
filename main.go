package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const LevelDebug = "DEBUG"
const LevelInfo = "INFO"
const LevelNotice = "NOTICE"
const LevelWarning = "WARNING"
const LevelError = "ERROR"
const LevelEmergency = "EMERGENCY"
const LevelFatal = "FATAL"

var LevelWeights = map[string]int{
	LevelDebug:     0,
	LevelInfo:      1,
	LevelNotice:    2,
	LevelWarning:   3,
	LevelError:     4,
	LevelEmergency: 5,
	LevelFatal:     6,
}

var levelWeight int

var LogDir = "./logs"
var logDirExists = false
var start = float64(0)
var lastStep = float64(0)

var IncludeRuntime = false
var IncludeStep = false

var LogRequestsSeparately = false
var HideRequestsFromMainLog = false

var minimumLogLevel = LevelNotice

var Component = ""

// init sets some default values by reading the environment variables.
// The following environment variables are supported:
// LOGGER_LOG_DIR: The directory where the log files are stored. Default: ./logs
// LOGGER_INCLUDE_RUNTIME: If set to true, the runtime is included in the log entry. Default: false
// LOGGER_INCLUDE_STEP: If set to true, the step is included in the log entry. Default: false
// LOGGER_LOG_REQUESTS_SEPARATELY: If set to true, the requests are logged in a separate file. Default: false
// LOGGER_HIDE_REQUESTS_FROM_MAIN_LOG: If set to true, the requests are not logged in the main log file. Default: false
func init() {
	logDirTemp, logDirIsSet := os.LookupEnv("LOGGER_LOG_DIR")
	if logDirIsSet {
		log.Println("LOGGER: Using log directory from environment variable: " + logDirTemp)
		logDirTemp = strings.TrimSpace(logDirTemp)
		if logDirTemp != "" {
			LogDir = logDirTemp
		}
	}

	includeRuntimeTemp, includeRuntimeIsSet := os.LookupEnv("LOGGER_INCLUDE_RUNTIME")
	if includeRuntimeIsSet {
		log.Println("LOGGER: Using include runtime from environment variable: " + includeRuntimeTemp)
		includeRuntimeTemp = strings.TrimSpace(includeRuntimeTemp)
		if includeRuntimeTemp == "true" {
			IncludeRuntime = true
		}
	}

	includeStepTemp, includeStepIsSet := os.LookupEnv("LOGGER_INCLUDE_STEP")
	if includeStepIsSet {
		log.Println("LOGGER: Using include step from environment variable: " + includeStepTemp)
		includeStepTemp = strings.TrimSpace(includeStepTemp)
		if includeStepTemp == "true" {
			IncludeStep = true
		}
	}

	logRequestsSeparatelyTemp, logRequestsSeparatelyIsSet := os.LookupEnv("LOGGER_LOG_REQUESTS_SEPARATELY")
	if logRequestsSeparatelyIsSet {
		log.Println("LOGGER: Using log requests separately from environment variable: " + logRequestsSeparatelyTemp)
		logRequestsSeparatelyTemp = strings.TrimSpace(logRequestsSeparatelyTemp)
		if logRequestsSeparatelyTemp == "true" {
			LogRequestsSeparately = true
		}
	}

	hideRequestsFromMainLogTemp, hideRequestsFromMainLogIsSet := os.LookupEnv("LOGGER_HIDE_REQUESTS_FROM_MAIN_LOG")
	if hideRequestsFromMainLogIsSet {
		log.Println("LOGGER: Using hide requests from main log from environment variable: " + hideRequestsFromMainLogTemp)
		hideRequestsFromMainLogTemp = strings.TrimSpace(hideRequestsFromMainLogTemp)
		if hideRequestsFromMainLogTemp == "true" {
			HideRequestsFromMainLog = true
		}
	}

	minimumLogLevelTemp, minimumLogLevelIsSet := os.LookupEnv("LOGGER_MINIMUM_LOG_LEVEL")
	if minimumLogLevelIsSet {
		log.Println("LOGGER: Using minimum log level from environment variable: " + minimumLogLevelTemp)
		minimumLogLevelTemp = strings.TrimSpace(minimumLogLevelTemp)
		if minimumLogLevelTemp != "" {
			log.Println("LOGGER: Setting minimum log level to: " + minimumLogLevelTemp)
			minimumLogLevelTemp = strings.ToUpper(minimumLogLevelTemp)
			for key := range LevelWeights {
				if key == minimumLogLevelTemp {
					minimumLogLevel = minimumLogLevelTemp
					break
				}
			}
		}
	}

	// check if logs directory exists, if not create it
	_, err := os.Stat(LogDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(LogDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("LOGGER: Created log directory: " + LogDir)
		logDirExists = true
	}

	// set level weights
	levelWeight = LevelWeights[minimumLogLevel]
}

func SetMinimumLogLevel(level string) {
	level = strings.ToUpper(level)
	found := false
	for key := range LevelWeights {
		if key == level {
			minimumLogLevel = level
			levelWeight = LevelWeights[level]
			found = true
			break
		}
	}

	if !found {
		SetMinimumLogLevel(LevelNotice)
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

func createHttpClient() *http.Client {
	// create http client
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return client
}

// l is the main logging function.
// It logs the given content to the main log file.
// It's internal and should not be used directly because we provide wrapper functions for each log level below.
func l(level string, content string) {
	// check if level is one of the supported levels
	if _, ok := LevelWeights[level]; !ok {
		log.Println("LOGGER: Invalid log level: " + level)
		return
	}

	// check if level is allowed
	if levelWeight > LevelWeights[level] {
		log.Println("LOGGER: Log level not allowed: " + level)
		log.Printf("LOGGER: Level weight of minimum log level: %d, level weight of selected level: %d\n", levelWeight, LevelWeights[level])
		return
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

	if Component != "" {
		entry += "[" + Component + "]"
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

	if level == LevelFatal {
		panic(content)
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
	if levelWeight > LevelWeights[LevelDebug] {
		log.Println("Debug mode is disabled. To enable it set the minimum log level to debug.")
		return
	}

	l(LevelDebug, content)
}

// DebugAsync logs a debug message asynchronously by calling logger.l as goroutine.
func DebugAsync(content string) {
	go Debug(content)
}

// Info logs an info message.
func Info(content string) {
	if levelWeight > LevelWeights[LevelInfo] {
		log.Println("Info mode is disabled. To enable it set the minimum log level to info.")
		return
	}

	l(LevelInfo, content)
}

// InfoAsync logs an info message asynchronously by calling logger.l as goroutine.
func InfoAsync(content string) {
	go Info(content)
}

// Warning logs a warning message.
func Warning(content string) {
	if levelWeight > LevelWeights[LevelWarning] {
		log.Println("Warning mode is disabled. To enable it set the minimum log level to warning.")
		return
	}

	l(LevelWarning, content)
}

// WarningAsync logs a warning message asynchronously by calling logger.l as goroutine.
func WarningAsync(content string) {
	go Warning(content)
}

// Error logs an err message.
func Error(content string) {
	if levelWeight > LevelWeights[LevelError] {
		log.Println("Error mode is disabled. To enable it set the minimum log level to error.")
		return
	}

	l(LevelError, content)
}

// ErrorAsync logs an err message asynchronously by calling logger.l as goroutine.
func ErrorAsync(content string) {
	go Error(content)
}

// Fatal logs a fatal message.
func Fatal(content string) {
	l(LevelFatal, content)
	log.Fatal(content)
}

// FatalAsync logs a fatal message asynchronously by calling logger.l as goroutine.
func FatalAsync(content string) {
	go Fatal(content)
}

// LogSimpleRequest logs a request.
// This is mainly used by Panorama.
// If HideRequestsFromMainLog is true, the request will not be logged to the main log file but only when LogRequestsSeparately is true.
func LogSimpleRequest(method string, path string, userAgent string, ip string) {
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
		f, err := os.OpenFile(LogDir+"/requests-simple-"+date+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

func in_array(val interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				return true
			}
		}
	}

	return false
}
