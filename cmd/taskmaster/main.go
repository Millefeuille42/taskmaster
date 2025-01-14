package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func handleCommands(command chan<- []string, shutdown chan<- os.Signal) {
	fmt.Println("Taskmaster TUI (Type 'help' for commands, 'exit' to quit)")
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			break
		}
		args := strings.Split(strings.TrimSpace(cmd), " ")

		slog.Debug("executing command", slog.String("command", args[0]))
		switch args[0] {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("help: Print this help")
			fmt.Println("exit: Shutdown taskmaster")
			for _, c := range commands {
				fmt.Println(c.String())
			}
		case "exit":
			fmt.Println("Exiting...")
			slog.Info("Received exit command, exiting...")
			shutdown <- syscall.SIGINT
			break
		default:
			command <- args
		}
	}
}

func parseFlags() func() error {
	logLevel := flag.String("loglevel", "info", "Set the logging level (debug, info, warn, error)")
	logFile := flag.String("logfile", "./taskmaster.log", "Optional file to write logs to")
	flag.Parse()

	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Invalid log level: %s\n", *logLevel)
		os.Exit(1)
	}

	var opts slog.HandlerOptions
	opts.Level = level

	var file *os.File
	var closer = func() error { return nil }
	if *logFile == "stdout" {
		file = os.Stdout
	} else if *logFile == "stderr" {
		file = os.Stderr
	} else {
		var err error
		file, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		closer = file.Close
	}
	logger := slog.New(slog.NewTextHandler(file, &opts))
	slog.SetDefault(logger)
	slog.Debug("parsing", slog.String("function", "parseFlags"), slog.String("logLevel", *logLevel))
	slog.Debug("parsing", slog.String("function", "parseFlags"), slog.String("logFile", *logFile))
	return closer
}

func main() {
	loggerClose := parseFlags()
	defer loggerClose()

	slog.Debug("parsing", slog.String("function", "main"), slog.Int("flag.NArg()", flag.NArg()))
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	slog.Info("Starting Taskmaster", slog.Int("pid", os.Getpid()))
	configs := parseConfig()

	command := make(chan []string)
	shutdown := make(chan os.Signal)
	reloadSignal := make(chan os.Signal)
	defer close(command)
	defer close(shutdown)
	defer close(reloadSignal)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reloadSignal, syscall.SIGHUP)

	go handleCommands(command, shutdown)
	slog.Info("Taskmaster started")
	programManager(configs, shutdown, reloadSignal, command)
}
