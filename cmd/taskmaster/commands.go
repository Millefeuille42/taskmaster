package main

import (
	"fmt"
	"log/slog"
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
}

func list(configs *map[string]Config, args []string) error {
	for _, config := range *configs {
		fmt.Println(config.String())
	}
	return nil
}

func reload(configs *map[string]Config, args []string) error {
	slog.Info("Reloading configuration")
	*configs = parseConfig()
	fmt.Println("Reloaded configuration")
	slog.Info("Reloaded configuration")
	return nil
}
