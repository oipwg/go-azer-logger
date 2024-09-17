package logger

import (
    "fmt"
    "os"
)

var (
    runtime *Runtime
    muted   = &OutputSettings{}
    verbose = &OutputSettings{
        Info:  true,
        Timer: true,
        Error: true,
    }
    // Add a variable to hold the log file pointer
    logFile *os.File
)

func init() {
    runtime = &Runtime{
        Writers: []OutputWriter{
            NewStandardOutput(os.Stderr),
        },
    }

    // Check for LOG_FILE environment variable
    logFilePath := os.Getenv("LOG_FILE")
    if logFilePath != "" {
        // Try to open the log file
        var err error
        logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", logFilePath, err)
            // If we can't open the file, continue without file logging
        } else {
            // Create a new StandardWriter for the file
            fileWriter := NewStandardOutput(logFile)
            fileWriter.(*StandardWriter).ColorsEnabled = false // Disable colors for file output

            // Hook the file writer into the runtime
            Hook(fileWriter)
        }
    }
}

// Add a function to close the log file when the application exits
func CloseLogFile() {
    if logFile != nil {
        logFile.Close()
    }
}

type OutputWriter interface {
    Init()
    Write(log *Log)
}

type OutputSettings struct {
	Info  bool
	Timer bool
	Error bool
}

type Runtime struct {
	Writers []OutputWriter
}

func (runtime *Runtime) Log(log *Log) {
	if len(runtime.Writers) == 0 {
		return
	}

	// Avoid getting into a loop if there is just one writer
	if len(runtime.Writers) == 1 {
		runtime.Writers[0].Write(log)
	} else {
		for _, w := range runtime.Writers {
			w.Write(log)
		}
	}
}

// Add a new writer
func Hook(writer OutputWriter) {
	writer.Init()
	runtime.Writers = append(runtime.Writers, writer)
}

// Legacy method
func SetOutput(file *os.File) {
	writer := NewStandardOutput(file)
	runtime.Writers[0] = writer
}
