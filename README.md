# Panorama Logger

> *This package is part of the [Panorama CMS](https://github.com/panorama-cms/panorama) project.*

## Usage

```go
package main

import (
    "github.com/panorama-cms/logger"
)

func main() {
	// The init function tries to read some values from the environment variables.
	// If you want to set them manually, you can do it like this:
	logger.IncludeRuntime = true // default: false; this will include the app runtime in the log message
	logger.IncludeStep = true // default: false; this will include the time since the last log message in the log message
	logger.LogRequestsSeparately = true // default: false; this will log requests in a separate file
	logger.HideRequestsFromMainLog = true // default: false; this will prevent requests from being logged in the main log file. Note, that this will only work if LogRequestsSeparately is set to true.
	logger.LogDir = "./logs" // default: "./logs"; this will set the directory where the log files will be stored
	
	// Log debugging information
	logger.Debug("Debugging information")
	
	// Log information
	logger.Info("Information")
	
	// Log a warning
	logger.Warning("Warning")
	
	// Log an error
	logger.Error("Error")
	
	// Log a fatal error
	// Note, that this will not end the application like log.Fatal does.
	logger.Fatal("Fatal error")
	
	// Log a request
	logger.LogRequest("GET", "/api/v1/users", "Some fancy-dancy user agent", "127.0.0.1")
}
```