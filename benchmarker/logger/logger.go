package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/yahoojapan/yisucon/benchmarker/config"
)

type Logger struct {
	LogFile *os.File
	wg      *sync.WaitGroup
}

var (
	logger *Logger
	once   sync.Once
)

func GetLogger() *Logger {
	var err error
	once.Do(func() {
		logger = &Logger{
			wg: new(sync.WaitGroup),
		}

		if _, err = os.Stat(config.LogFilePath); err != nil {
			if _, err = os.Stat(filepath.Dir(config.LogFilePath)); err != nil {
				os.MkdirAll(filepath.Dir(config.LogFilePath), 0755)
			}
			file, err := os.Create(config.LogFilePath)
			if err != nil {
				return
			}
			file.Close()
		}

		logger.LogFile, err = os.OpenFile(config.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)

	})

	if err == nil {
		log.SetOutput(io.MultiWriter(logger.LogFile, os.Stdout))
	}

	log.SetFlags(log.Ldate | log.Ltime)

	return logger
}

func (l Logger) Println(val interface{}) {
	l.wg.Add(1)
	go func() {
		log.Println(val)
		l.wg.Done()
	}()
}

func (l Logger) Printf(format string, val interface{}) {
	l.wg.Add(1)
	go func() {
		log.Printf(format, val)
		l.wg.Done()
	}()
}

func (l Logger) Fatalln(val interface{}) {
	l.wg.Add(1)
	go func() {
		log.Fatalln(val)
		l.wg.Done()
	}()
}

func (l *Logger) Close() error {
	l.wg.Wait()
	return l.LogFile.Close()
}
