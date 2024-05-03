package main

import (
	"flag"
	"log/slog"
	"os"
)

var version string

func printVersion() {
	if version == "" {
		version = "development"
	}

	println("amock version " + version)
}

var DebugValue = false

var versionFlag = flag.Bool("version", false, "Print the current version and exit")
var helpFlag = flag.Bool("help", false, "Print help message and exit")
var debugFlag = flag.Bool("debug", false, "Enable debug logging")

func parseFlags() {
	flag.BoolVar(versionFlag, "v", false, "Print the current version and exit")
	flag.BoolVar(debugFlag, "d", false, "Enable debug logging")
	flag.BoolVar(helpFlag, "h", false, "Print help message and exit")
	DebugValue = *debugFlag

	flag.Parse()

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	if *helpFlag || len(flag.Args()) >= 0 && flag.Arg(0) == "help" {
		println("Usage:\n\tamock [host:port] [flags]")
		println("\n[host:port] - (optional) The host and port to bind the server to")
		println("\nFlags: (optional)")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if DebugValue {
		LogLevel.Set(slog.LevelDebug)
		slog.SetLogLoggerLevel(LogLevel.Level())
	}
}
