package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	Log *log.Logger
)

func Init(logFilePath string) error {
	dir := filepath.Dir(logFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	Log = log.New(multiWriter, "", 0)
	return nil
}

func LogError(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[ERROR] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}

func LogInfo(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[INFO] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}

func LogDebug(format string, v ...interface{}) {
	if os.Getenv("DEBUG") != "true" {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[DEBUG] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}

func LogWarn(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[WARN] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}
