package logx

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

const (
	maxSizeBytes = 1024 * 1024
	maxBackups   = 7
)

type Logger struct {
	mu        sync.Mutex
	logger    *log.Logger
	file      *os.File
	path      string
	useStdout bool
}

func New(path string) *Logger {
	l := &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		path:   path,
	}
	if path == "" {
		l.useStdout = true
		registerLogger(l)
		return l
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.useStdout = true
		registerLogger(l)
		return l
	}
	l.file = file
	l.logger.SetOutput(file)
	registerLogger(l)
	return l
}

func (l *Logger) Info(msg string) {
	l.log("", msg)
}

func (l *Logger) Error(msg string) {
	l.log("ERROR: ", msg)
}

func (l *Logger) Reopen() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.path == "" {
		return
	}
	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.useStdout = true
		l.logger.SetOutput(os.Stdout)
		if l.file != nil {
			_ = l.file.Close()
			l.file = nil
		}
		return
	}
	if l.file != nil {
		_ = l.file.Close()
	}
	l.file = file
	l.useStdout = false
	l.logger.SetOutput(file)
}

func (l *Logger) log(prefix string, msg string) {
	rotated := false
	l.mu.Lock()
	if l.rotateIfNeededLocked() {
		rotated = true
	}
	if prefix == "" {
		l.logger.Println(msg)
	} else {
		l.logger.Println(prefix + msg)
	}
	l.mu.Unlock()
	if rotated {
		sendSIGHUP()
	}
}

func (l *Logger) rotateIfNeededLocked() bool {
	if l.path == "" {
		return false
	}
	info, err := os.Stat(l.path)
	if err != nil || info.Size() < maxSizeBytes {
		return false
	}

	rotateMu.Lock()
	defer rotateMu.Unlock()

	info, err = os.Stat(l.path)
	if err != nil || info.Size() < maxSizeBytes {
		return false
	}

	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}

	_ = os.Remove(l.path + "." + strconv.Itoa(maxBackups))
	for i := maxBackups - 1; i >= 1; i-- {
		src := l.path + "." + strconv.Itoa(i)
		dst := l.path + "." + strconv.Itoa(i+1)
		if fileExists(src) {
			_ = os.Rename(src, dst)
		}
	}
	_ = os.Rename(l.path, l.path+".1")

	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.useStdout = true
		l.logger.SetOutput(os.Stdout)
		return true
	}
	l.file = file
	l.useStdout = false
	l.logger.SetOutput(file)
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var (
	registryMu sync.Mutex
	registry   []*Logger
	signalOnce sync.Once
	rotateMu   sync.Mutex
)

func registerLogger(logger *Logger) {
	registryMu.Lock()
	registry = append(registry, logger)
	registryMu.Unlock()
	signalOnce.Do(setupSIGHUPHandler)
}

func setupSIGHUPHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for range sigs {
			ReopenAll()
		}
	}()
}

func ReopenAll() {
	registryMu.Lock()
	loggers := make([]*Logger, len(registry))
	copy(loggers, registry)
	registryMu.Unlock()

	for _, logger := range loggers {
		if logger == nil {
			continue
		}
		logger.Reopen()
	}
}

func sendSIGHUP() {
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		return
	}
	_ = proc.Signal(syscall.SIGHUP)
}
