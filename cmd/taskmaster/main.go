package main

import (
	"bufio"
	"errors"
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

func handleCommands(
	configs *map[string]Config,
	output chan<- string,
	shutdown chan os.Signal,
) {
	reloadSignal := make(chan os.Signal)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	defer close(reloadSignal)

	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-shutdown:
			return
		case <-reloadSignal:
			*configs = parseConfig()
		default:
			output <- "> "
			cmd, _ := reader.ReadString('\n')
			args := strings.Split(strings.TrimSpace(cmd), " ")
			cmd = args[0]
			switch cmd {
			case "help":
				output <- "Available commands:\n"
				for _, command := range commands {
					output <- command.String() + "\n"
				}
			case "exit":
				output <- "Exiting...\n"
				shutdown <- os.Interrupt
			default:
				if command, ok := commands[cmd]; ok {
					err := command.Function(configs, args, output)
					if err != nil {
						slog.Error(err.Error())
					}
					continue
				}
				output <- "Unknown command: " + cmd + "\n"
			}
		}
	}
}

func tui(configs map[string]Config) {
	output := make(chan string)
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	defer close(output)
	defer close(shutdown)

	fmt.Println("Taskmaster TUI (Type 'help' for commands, 'exit' to quit)")

	go handleCommands(&configs, output, shutdown)
	for {
		select {
		case msg := <-output:
			fmt.Print(msg)
		case <-shutdown:
			return
		}
	}
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	if len(os.Args) <= 1 {
		slog.Error(fmt.Sprintf("usage: %s [file|directory]...", os.Args[0]))
		os.Exit(1)
	}

	configs := parseConfig()
	tui(configs)
}
