package main

import (
	"fmt"
)

type CommandFunction func(*map[string]Config, []string, chan<- string) error

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

func list(configs *map[string]Config, args []string, output chan<- string) error {
	for _, config := range *configs {
		output <- config.String() + "\n"
	}
	return nil
}

func reload(configs *map[string]Config, args []string, output chan<- string) error {
	*configs = parseConfig()
	output <- "Reloaded config\n"
	return nil
}
