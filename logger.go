package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jwalton/gchalk"
)

func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.SetFlags(log.Ldate | log.Ltime)
		log.SetPrefix("[amock]: ")
		remoteAddr := gchalk.Bold(r.RemoteAddr)
		method := requestMethodColor(r.Method)
		start := time.Now()

		handler.ServeHTTP(w, r)

		elapsed := time.Since(start).String()
		elapsed = gchalk.WithItalic().Dim("(" + elapsed + ")")

		// [amock]: 2024/04/01 02:43:10 - localhost:8000 | 127.0.0.1:12345 -> GET /api/v1/users (1.234s)
		log.Printf("- %s | %s -> %s %s %s", r.Host, remoteAddr, method, r.URL, elapsed)
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
