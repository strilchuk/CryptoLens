package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
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
	Log = log.New(multiWriter, "", log.LstdFlags)
	return nil
}
