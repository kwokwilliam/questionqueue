package handler

import (
	"log"
	"net/http"
	"time"
)

type Logger struct {
	handler http.Handler
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.handler.ServeHTTP(w, r)

	log.Printf("%s %s %v", r.Method, r.URL.String(), time.Since(start))
}

func NewLogger(h http.Handler) *Logger {
	return &Logger{h}
}
