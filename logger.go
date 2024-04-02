package main

import (
	"flag"
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

var Verbose = false
var verboseFlag = flag.Bool("verbose", false, "Enable verbose logging")

func InitLogger() {
	flag.BoolVar(verboseFlag, "v", false, "Shorthand for `--verbose`")
	flag.Parse()
	Verbose = *verboseFlag

	if Verbose {
		LogLevel.Set(slog.LevelDebug)
		slog.SetLogLoggerLevel(LogLevel.Level())
	}
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.SetFlags(log.Ldate | log.Ltime)
		log.SetPrefix("[amock]: ")

		remoteAddr := gchalk.Bold(r.RemoteAddr)
		method := requestMethodColor(r.Method)
		recorder := &StatusRecorder{w, http.StatusOK}

		next.ServeHTTP(recorder, r)

		elapsed := time.Since(start).String()
		elapsed = gchalk.WithItalic().Dim("(" + elapsed + ")")
		status := getStatusColor(recorder.Status)

		// [amock]: 2024/04/01 02:43:10 - localhost:8000 | 127.0.0.1:12345 -> GET /api/v1/users - 200 OK (1.234s)
		log.Printf("- %s | %s -> %s %s - %s %s", r.Host, remoteAddr, method, r.URL, status, elapsed)
	})
}

func requestMethodColor(m string) string {
	var method string
	switch m {
	case http.MethodGet:
		method = gchalk.WithBrightWhite().WithBold().BgBrightBlue(" " + m + " ")
	case http.MethodPost:
		method = gchalk.WithBrightWhite().WithBold().BgBrightGreen(" " + m + " ")
	case http.MethodPut:
		method = gchalk.WithBrightWhite().WithBold().BgBrightYellow(" " + m + " ")
	case http.MethodPatch:
		method = gchalk.WithBrightWhite().WithBold().BgBrightCyan(" " + m + " ")
	case http.MethodDelete:
		method = gchalk.WithBrightWhite().WithBold().BgRed(" " + m + " ")
	case http.MethodOptions:
		method = gchalk.WithBrightWhite().WithBold().BgBlue(" " + m + " ")
	case http.MethodHead:
		method = gchalk.WithBrightWhite().WithBold().BgMagenta(" " + m + " ")
	default:
		method = m
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
