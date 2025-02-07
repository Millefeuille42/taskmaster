package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sync"
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

func list(
	statusCommand chan<- string,
	_ chan<- Config,
	configs *map[string]Config,
	_ []string,
) error {
	for _, config := range *configs {
		fmt.Println(config.String())
	}
	statusCommand <- "done"
	return nil
}

// c'est bourin mais ça marche :^)
func config_reconcilier(config Config, old_config Config, configChannel chan<- Config) {
	if !reflect.DeepEqual(config.Command, old_config.Command) ||
	   config.NumProcs < old_config.NumProcs ||
	   config.Stdout != old_config.Stdout ||
	   config.Stderr != old_config.Stderr ||
	   !reflect.DeepEqual(config.Env, old_config.Env) ||
	   config.WorkDir != old_config.WorkDir ||
	   config.Umask != old_config.Umask {
		stopProgram(old_config, configChannel)
		startProc(config, configChannel)
		return
	}
	if config.NumProcs > old_config.NumProcs {
		startProc(config, configChannel)
	}
}

func reload(
	statusCommand chan<- string,
	configChannel chan<- Config,
	configs *map[string]Config,
	_ []string,
) error {
	slog.Info("Reloading configuration")

	old_configs := *configs
	*configs = parseConfig()

	for name, config := range old_configs {
		if (*configs)[name].name == "" {
			// delete process
			stopProgram(config, configChannel)
		}
	}

	for name, config := range *configs {
		if old_configs[name].name == "" {
			// start process
			if config.AutoStart {
				startProc(config, configChannel)
			}
		} else {
			slog.Info("AAAAAAAAAA")
			// TODO: ca fonctionne po :c
			// currently when the program reload it loose track of all the pids
			// thought i would like try to copy them or sth, but doesnt work much
			config.pids = clonePidsMap(old_configs[name].pids)
			config_reconcilier(config, old_configs[name], configChannel)
			// delete(*old_config, name)
		}
	}

	fmt.Println("Reloaded configuration")
	slog.Info("Reloaded configuration")
	statusCommand <- "done"
	return nil
}

func ps(
	statusCommand chan<- string,
	_ chan<- Config,
	configs *map[string]Config,
	_ []string,
) error {
	for _, config := range *configs {
		if len(config.pids) <= 0 {
			continue
		}
		printProgramStatus(config)
	}
	statusCommand <- "done"
	return nil
}

func stat(
	statusCommand chan<- string,
	_ chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	if len(args) < 2 {
		error_msg := "usage: status <program>"
		statusCommand <- error_msg
		return errors.New(error_msg)
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

	statusCommand <- "done"
	return nil
}

func start(
	statusCommand chan<- string,
	configChannel chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	if len(args) < 2 {
		error_msg := "usage: start <program>"
		statusCommand <- error_msg
		return errors.New(error_msg)
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

	statusCommand <- "done"
	return nil
}

func stop(
	statusCommand chan<- string,
	configChannel chan<- Config,
	configs *map[string]Config,
	args []string,
) error {
	if len(args) < 2 {
		error_msg := "usage: stop <program>"
		statusCommand <- error_msg
		return errors.New(error_msg)
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

	statusCommand <- "done"
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
		error_msg := "usage: restart <program>"
		statusCommand <- error_msg
		return errors.New(error_msg)
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

	return start(statusCommand, configChannel, configs, args)
}
