package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func handleCommand(cmd string, configs *map[string]Config) error {
	args := strings.Split(strings.TrimSpace(cmd), " ")
	cmd = args[0]
	slog.Debug("executing command", slog.String("command", cmd))
	switch cmd {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("help: Print this help")
		fmt.Println("exit: Shutdown " + os.Args[0])
		for _, command := range commands {
			fmt.Println(command.String())
		}
	case "exit":
		fmt.Println("Exiting...")
		slog.Info("Received exit command, exiting...")
		return os.ErrProcessDone
	default:
		if command, ok := commands[cmd]; ok {
			return command.Function(configs, args)
		}
		fmt.Println("Unknown command: " + cmd)
	}

	return nil
}

func tui(configs map[string]Config) {
	input := make(chan string)
	shutdown := make(chan os.Signal)
	reloadSignal := make(chan os.Signal)
	defer close(input)
	defer close(shutdown)
	defer close(reloadSignal)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reloadSignal, syscall.SIGHUP)

	fmt.Println("Taskmaster TUI (Type 'help' for commands, 'exit' to quit)")
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			cmd, _ := reader.ReadString('\n')
			err := handleCommand(cmd, &configs)
			if err != nil {
				if errors.Is(err, os.ErrProcessDone) {
					shutdown <- syscall.SIGINT
					return
				}
				slog.Error(err.Error())
			}
		}
	}()

	for {
		select {
		case <-shutdown:
			slog.Debug("Received shutdown signal")
			return
		case <-reloadSignal:
			slog.Debug("Received reload signal")
			// TODO Make config thread safe
			configs = parseConfig()
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
	slog.Info("Taskmaster started")
	tui(configs)
}
