package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func handleCommands(
	command chan<- []string,
	shutdown chan<- os.Signal,
	statusCommand <-chan string,
) {
	fmt.Println("Taskmaster TUI (Type 'help' for commands, 'exit' to quit)")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$> ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
			} else {
				_, _ = fmt.Fprintln(os.Stderr, "error:", err)
			}
			shutdown <- syscall.SIGINT
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
			msg := <-statusCommand
			if msg != "done" {
				fmt.Println(msg)
			}
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
	shutdown := make(chan os.Signal, 1)
	statusCommand := make(chan string)
	reloadSignal := make(chan os.Signal, 1)
	defer close(command)
	defer close(shutdown)
	defer close(reloadSignal)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reloadSignal, syscall.SIGHUP)

	populateCommands(reloadSignal)
	go handleCommands(command, shutdown, statusCommand)
	slog.Info("Taskmaster started")
	programManager(configs, shutdown, reloadSignal, command, statusCommand)
}
