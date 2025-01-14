package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func programManager(
	configs map[string]Config,
	shutdown <-chan os.Signal,
	reloadSignal <-chan os.Signal,
	command <-chan []string,
) {
	configChannel := make(chan Config)

	for _, config := range configs {
		if config.AutoStart {
			startProc(config, configChannel)
		}
	}
	for {
		select {
		case <-shutdown:
			slog.Debug("Received shutdown signal")
			return
		case <-reloadSignal:
			slog.Debug("Received reload signal")
			slog.Info("Reloading configuration")
			configs = parseConfig()
			fmt.Println("Reloaded configuration")
			slog.Info("Reloaded configuration")
			// TODO Run new programs, shutdown old ones, etc
		case args := <-command:
			if cmd, ok := commands[args[0]]; ok {
				err := cmd.Function(configChannel, &configs, args)
				if err != nil {
					slog.Error(err.Error(), slog.String("command", args[0]))
				}
				continue
			}
			fmt.Println("Unknown command: " + args[0])
		case config := <-configChannel:
			pids := config.pids
			for pid, status := range pids {
				if status.Running != false {
					continue
				}
				delete(config.pids, pid)
				if status.ExitCode == -1 {
					// This means it has been stopped by a signal
					//  thus it is highly possible that it has been shutdown
					//  with the stop command
					continue
				}
				if status.ExitedEarly || !isExitCodeValid(config, status) {
					if config.RestartWhen == "unexpected" {
						startProc(config, configChannel)
						continue
					}
				}
				if config.RestartWhen == "always" {
					startProc(config, configChannel)
				}
			}
			configs[config.name] = config
		}
	}
}

func runProgram(config Config, configChannel chan<- Config) error {
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
	pid := cmd.Process.Pid
	config.pids[pid] = ProgramStatus{
		Running:     true,
		ExitedEarly: false,
		ExitCode:    0,
	}
	configChannel <- config

	startTime := time.Now().Add(config.StartTime)
	err = cmd.Wait()
	config.pids[pid] = ProgramStatus{
		Running:     false,
		ExitedEarly: time.Now().Before(startTime),
		ExitCode:    cmd.ProcessState.ExitCode(),
	}
	if err != nil {
		slog.Error(fmt.Sprintf("%s: exited with error: %s", config.Command[0], err.Error()))
		configChannel <- config
	}
	configChannel <- config
	slog.Info(fmt.Sprintf("%s: exited successfully", config.Command[0]))
	return nil
}

func stopProgram(config Config, _ chan<- Config) error {
	if len(config.pids) == 0 {
		slog.Error(
			"Could not stop program, no running instances",
			slog.String("program", config.name),
		)
		return nil
	}

	stopSignal, err := stringToSignal(config.StopSignal)
	if err != nil {
		return err
	}

	for pid := range config.pids {
		slog.Debug(
			"stop",
			slog.String("program", config.name),
			slog.String("signal", stopSignal.String()),
			slog.Int("pid", pid),
		)
		err = syscall.Kill(pid, stopSignal)
		if err != nil {
			slog.Error("stop",
				slog.String("error", err.Error()),
				slog.String("program", config.name),
				slog.String("signal", stopSignal.String()),
				slog.Int("pid", pid),
			)
		}
		// TODO Wait for stop_time then kill
	}
	return nil
}
