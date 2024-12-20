package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

type CommandFunction func(chan<- Config, *map[string]Config, []string) error

type Command struct {
	Name     string
	HelpText string
	Function CommandFunction
}

func (c *Command) String() string {
	return fmt.Sprintf("%s: %s", c.Name, c.HelpText)
}

var commands = map[string]Command{
	"list": {
		Name:     "list",
		HelpText: "List programs configured and their options",
		Function: list,
	},
	"reload": {
		Name:     "reload",
		HelpText: "Reload configuration",
		Function: reload,
	},
	"start": {
		Name:     "start",
		HelpText: "Start programs",
		Function: start,
	},
	"stop": {
		Name:     "stop",
		HelpText: "Stop programs",
		Function: stop,
	},
	"status": {
		Name:     "status",
		HelpText: "Get status of a program",
		Function: stat,
	},
	"ps": {
		Name:     "ps",
		HelpText: "List running programs",
		Function: ps,
	},
}

func list(_ chan<- Config, configs *map[string]Config, _ []string) error {
	for _, config := range *configs {
		fmt.Println(config.String())
	}
	return nil
}

func reload(_ chan<- Config, configs *map[string]Config, _ []string) error {
	slog.Info("Reloading configuration")
	*configs = parseConfig()
	fmt.Println("Reloaded configuration")
	slog.Info("Reloaded configuration")
	return nil
}

func printProgramStatus(config Config) {
	fmt.Printf("%s %s:\n", config.name, config.Command)
	for pid, status := range config.pids {
		programStatus := "running"
		if !status.Running {
			programStatus = "exited"
			if status.ExitedEarly {
				programStatus += " early"
			}
			programStatus += fmt.Sprintf(": %d", status.ExitCode)
		}
		fmt.Printf("\t%d (%s)\n", pid, programStatus)
	}
}

func ps(_ chan<- Config, configs *map[string]Config, _ []string) error {
	for _, config := range *configs {
		if len(config.pids) <= 0 {
			continue
		}
		printProgramStatus(config)
	}
	return nil
}

func stat(_ chan<- Config, configs *map[string]Config, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: status <program>")
	}
	for _, arg := range args[1:] {
		config, ok := (*configs)[arg]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "unknown program: %s\n", arg)
			continue
		}
		if len(config.pids) <= 0 {
			fmt.Println("not running")
			continue
		}
		printProgramStatus(config)
	}

	return nil
}

func start(configChannel chan<- Config, configs *map[string]Config, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: start <program>")
	}
	for _, arg := range args[1:] {
		config, ok := (*configs)[arg]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "unknown program: %s\n", arg)
			continue
		}
		go func() {
			err := runProgram(config, configChannel)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	return nil
}

func stop(configChannel chan<- Config, configs *map[string]Config, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: stop <program>")
	}

	for _, arg := range args[1:] {
		config, ok := (*configs)[arg]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "unknown program: %s\n", arg)
			continue
		}
		go func() {
			slog.Debug("stop", slog.String("program", arg))
			err := stopProgram(config, configChannel)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	return nil
}
