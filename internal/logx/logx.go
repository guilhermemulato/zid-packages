package logx

import (
	"log"
	"os"
	"sync"
)

type Logger struct {
	mu     sync.Mutex
	logger *log.Logger
}

func New(path string) *Logger {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &Logger{logger: log.New(os.Stdout, "", log.LstdFlags)}
	}
	return &Logger{logger: log.New(file, "", log.LstdFlags)}
}

func (l *Logger) Info(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println("ERROR: " + msg)
}
