package main

import (
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jwalton/gchalk"
)

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (sr *StatusRecorder) WriteHeader(status int) {
	sr.Status = status
	sr.ResponseWriter.WriteHeader(status)
}

var LogLevel = new(slog.LevelVar)

func Warn(msg string, args ...any) {
	slog.Warn(gchalk.Yellow(msg), args...)
}

func Error(msg string, args ...any) {
	slog.Error(gchalk.Red(msg), args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(gchalk.Dim(msg), args...)
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.SetFlags(log.Ldate | log.Ltime)
		log.SetPrefix("[amock]: ")

		remoteAddr := gchalk.Bold(r.RemoteAddr)
		method := RequestMethodColor(r.Method, true)
		recorder := &StatusRecorder{w, http.StatusOK}

		next.ServeHTTP(recorder, r)

		elapsed := time.Since(start).String()
		elapsed = gchalk.WithItalic().Dim("(" + elapsed + ")")
		status := getStatusColor(recorder.Status)

		// [amock]: 2024/04/01 02:43:10 - localhost:8000 | 127.0.0.1:12345 -> GET /api/v1/users - 200 OK (1.234s)
		log.Printf("- %s | %s -> %s %s - %s %s", r.Host, remoteAddr, method, r.URL, status, elapsed)
	})
}

func RequestMethodColor(m string, inverse bool) string {
	method := m

	if inverse {
		method = " " + m + " "
	}

	switch m {
	case http.MethodGet:
		method = gchalk.WithBold().BrightBlue(method)
	case http.MethodPost:
		method = gchalk.WithBold().BrightGreen(method)
	case http.MethodPut:
		method = gchalk.WithBold().BrightYellow(method)
	case http.MethodPatch:
		method = gchalk.WithBold().BrightCyan(method)
	case http.MethodDelete:
		method = gchalk.WithBold().Red(method)
	case http.MethodOptions:
		method = gchalk.WithBold().Blue(method)
	case http.MethodHead:
		method = gchalk.WithBold().Magenta(method)
	default:
		method = m
	}

	if inverse {
		method = gchalk.BgBrightWhite(method)
		method = gchalk.Inverse(method)
	}

	return method
}

func getStatusColor(status int) string {
	var color string
	switch {
	case status >= 200 && status < 300:
		color = gchalk.BrightGreen(gchalk.Bold(strconv.Itoa(status)), http.StatusText(status))
	case status >= 300 && status < 400:
		color = gchalk.BrightBlue(gchalk.Bold(strconv.Itoa(status)), http.StatusText(status))
	case status >= 400 && status < 500:
		color = gchalk.BrightYellow(gchalk.Bold(strconv.Itoa(status)), http.StatusText(status))
	case status >= 500:
		color = gchalk.BrightRed(gchalk.Bold(strconv.Itoa(status)), http.StatusText(status))
	default:
		color = gchalk.Bold(strconv.Itoa(status)) + http.StatusText(status)
	}
	return color
}
