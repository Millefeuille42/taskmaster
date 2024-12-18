package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func runProgram(config Config) error {
	if config.Command == nil || len(config.Command) <= 0 {
		return errors.New("no command provided")
	}

	var args []string
	if len(config.Command) > 1 {
		args = config.Command[1:]
	}
	cmd := exec.Command(config.Command[0], args...)
	if cmd.Err != nil {
		return cmd.Err
	}

	if config.Stdout != "" {
		stdout, err := os.OpenFile(config.Stdout, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer stdout.Close()
		cmd.Stdout = stdout
	}

	if config.Stderr != "" {
		stderr, err := os.OpenFile(config.Stderr, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer stderr.Close()
		cmd.Stderr = stderr
	}

	cmd.Dir = config.WorkDir
	if config.Env != nil {
		cmd.Env = os.Environ()
		for key, value := range config.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	slog.Debug(cmd.String())

	syscall.Umask(config.Umask)
	err := cmd.Start()
	if err != nil {
		return err
	}
	wait := make(chan error)
	go func(chan error) {
		wait <- cmd.Wait()
	}(wait)

	startTime := time.NewTimer(config.StartTime * time.Second)
	defer startTime.Stop()

	select {
	case err = <-wait:
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("%s: exited successfully", config.Command[0]))
	case <-startTime.C:
		if cmd.ProcessState.Exited() {
			slog.Error(fmt.Sprintf("%s: started successfully", config.Command[0]))
		} else {
			slog.Info(fmt.Sprintf("%s: started successfully", config.Command[0]))
		}
	}

	return nil
}

func tui(configs map[string]Config) {
	shutdown := make(chan os.Signal)
	defer close(shutdown)
	reloadSignal := make(chan os.Signal)
	defer close(reloadSignal)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Taskmaster TUI (Type 'help' for commands, 'exit' to quit)")

	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-shutdown:
			return
		case <-reloadSignal:
			configs = parseConfig()
		default:
			fmt.Print("> ")
			cmd, _ := reader.ReadString('\n')
			args := strings.Split(strings.TrimSpace(cmd), " ")
			cmd = args[0]
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
				shutdown <- os.Interrupt
			default:
				if command, ok := commands[cmd]; ok {
					err := command.Function(&configs, args)
					if err != nil {
						slog.Error(err.Error())
					}
					continue
				}
				fmt.Println("Unknown command: " + cmd)
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

	slog.Info("Starting Taskmaster")
	configs := parseConfig()
	slog.Info("Taskmaster started")
	tui(configs)
}
