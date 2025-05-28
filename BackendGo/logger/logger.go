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

// LogError логирует ошибку с информацией о файле и строке
func LogError(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[ERROR] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}

// LogInfo логирует информационное сообщение
func LogInfo(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[INFO] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}

// LogDebug логирует отладочное сообщение
func LogDebug(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	Log.Printf("[DEBUG] [%s:%d] %s", file, line, fmt.Sprintf(format, v...))
}
