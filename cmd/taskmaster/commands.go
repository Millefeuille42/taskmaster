package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

type CommandFunction func(*map[string]Config, []string) error

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
}

func list(configs *map[string]Config, _ []string) error {
	for _, config := range *configs {
		fmt.Println(config.String())
	}
	return nil
}

func reload(configs *map[string]Config, _ []string) error {
	slog.Info("Reloading configuration")
	*configs = parseConfig()
	fmt.Println("Reloaded configuration")
	slog.Info("Reloaded configuration")
	return nil
}

func start(configs *map[string]Config, args []string) error {
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
			err := runProgram(&config)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
			(*configs)[arg] = config
		}()
	}

	return nil
}

func stop(configs *map[string]Config, args []string) error {
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
			err := stopProgram(&config)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
			(*configs)[arg] = config
		}()
	}

	return nil
}
