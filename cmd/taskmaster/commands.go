package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"syscall"
	"time"
)

type CommandFunction func(chan<- string, chan<- Config, *map[string]Config, []string) error

type Command struct {
	Name     string
	HelpText string
	Function CommandFunction
}

func (c *Command) String() string {
	return fmt.Sprintf("%s: %s", c.Name, c.HelpText)
}

var commands map[string]Command

func populateCommands(reloadSignal chan<- os.Signal) {
	commands = map[string]Command{
		"list": {
			Name:     "list",
			HelpText: "List programs configured and their options",
			Function: list,
		},
		"reload": {
			Name:     "reload",
			HelpText: "Reload configuration",
			Function: reloadFactory(reloadSignal),
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
		"restart": {
			Name:     "restart",
			HelpText: "Restart programs",
			Function: restart,
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
}

func list(
	statusCommand chan<- string,
	_ chan<- Config,
	configs *map[string]Config,
	_ []string,
) error {
	for _, config := range *configs {
		fmt.Println(config.String())
	}
	statusCommand <- CommandDone
	return nil
}

func reloadFactory(reloadSignal chan<- os.Signal) CommandFunction {
	return func(
		statusCommand chan<- string,
		_ chan<- Config,
		_ *map[string]Config,
		_ []string,
	) error {
		reloadSignal <- syscall.SIGHUP
		statusCommand <- CommandDone
		return nil
	}
}

func ps(
	statusCommand chan<- string,
	_ chan<- Config,
	configs *map[string]Config,
	_ []string,
) error {
	for _, config := range *configs {
		config.lock.Lock()
		if len(config.pids) <= 0 {
			config.lock.Unlock()
			continue
		}
		config.lock.Unlock()
		printProgramStatus(config)
	}
	statusCommand <- CommandDone
	return nil
}

func stat(
	statusCommand chan<- string,
	_ chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	if len(args) < 2 {
		errMsg := "usage: status <program>"
		statusCommand <- errMsg
		return errors.New(errMsg)
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

	statusCommand <- CommandDone
	return nil
}

func start(
	statusCommand chan<- string,
	configChannel chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	if len(args) < 2 {
		errMsg := "usage: start <program>"
		statusCommand <- errMsg
		return errors.New(errMsg)
	}
	for _, arg := range args[1:] {
		config, ok := (*configs)[arg]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "unknown program: %s\n", arg)
			continue
		}
		config.restart = 0
		startProc(config, configChannel)
	}

	statusCommand <- CommandDone
	return nil
}

func stop(
	statusCommand chan<- string,
	configChannel chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	if len(args) < 2 {
		errMsg := "usage: stop <program>"
		statusCommand <- errMsg
		return errors.New(errMsg)
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

	statusCommand <- CommandDone
	return nil
}

func restart(
	statusCommand chan<- string,
	configChannel chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	var waitGroup sync.WaitGroup

	if len(args) < 2 {
		errMsg := "usage: restart <program>"
		statusCommand <- errMsg
		return errors.New(errMsg)
	}

	for _, arg := range args[1:] {
		config, ok := (*configs)[arg]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "unknown program: %s\n", arg)
			continue
		}
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			slog.Debug("stop", slog.String("program", arg))
			err := stopProgram(config, configChannel)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	waitGroup.Wait()

	for _, arg := range args[1:] {
		config, ok := (*configs)[arg]
		time.Sleep(config.StopTime)
		if !ok {
			continue
		}
		config.restart = 0
		startProc(config, configChannel)
	}

	statusCommand <- CommandDone
	return nil
}
