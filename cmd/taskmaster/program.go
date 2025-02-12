package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"time"
)

func processConfigDiff(config Config, oldConfig Config, configChannel chan<- Config) {
	// c'est bourin mais ça marche :^)
	if !reflect.DeepEqual(config.Command, oldConfig.Command) ||
		config.NumProcs < oldConfig.NumProcs ||
		config.Stdout != oldConfig.Stdout ||
		config.Stderr != oldConfig.Stderr ||
		!reflect.DeepEqual(config.Env, oldConfig.Env) ||
		config.WorkDir != oldConfig.WorkDir ||
		config.Umask != oldConfig.Umask {
		if err := stopProgram(oldConfig, configChannel); err != nil {
			slog.Error("An error occurred while stopping program",
				slog.String("error", err.Error()),
				slog.String("program", config.name),
			)
		}
		startProc(config, configChannel)
		return
	}
	if config.NumProcs > oldConfig.NumProcs {
		startProc(config, configChannel)
	}
}

func handleReload(
	configs *map[string]Config,
	configChannel chan<- Config,
) (map[string]Config, error) {
	oldConfigs := *configs
	*configs = parseConfig()

	for name, config := range oldConfigs {
		if (*configs)[name].name == "" {
			// delete process
			if err := stopProgram(config, configChannel); err != nil {
				slog.Error("An error occurred while stopping program",
					slog.String("error", err.Error()),
					slog.String("program", config.name),
				)
			}
		}
	}

	for name, config := range *configs {
		if oldConfigs[name].name == "" {
			// start process
			if config.AutoStart {
				startProc(config, configChannel)
			}
		} else {
			slog.Info("AAAAAAAAAA")
			// TODO: ca fonctionne po :c
			// currently when the program reload it loose track of all the pids
			// thought i would like try to copy them or sth, but doesnt work much
			config.pids = clonePidsMap(oldConfigs[name].pids)
			processConfigDiff(config, oldConfigs[name], configChannel)
			// delete(*old_config, name)
		}
	}

	return nil, nil
}

func programManager(
	configs map[string]Config,
	shutdown <-chan os.Signal,
	reloadSignal <-chan os.Signal,
	command <-chan []string,
	statusCommand chan<- string,
) {
	var err error
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
			configs, err = handleReload(&configs, configChannel)
			fmt.Println("Reloaded configuration")
			slog.Info("Reloaded configuration")
		case args := <-command:
			if cmd, ok := commands[args[0]]; ok {
				err = cmd.Function(statusCommand, configChannel, &configs, args)
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
				if config.restart >= config.MaxRestarts {
					continue
				} else {
					config.restart += 1
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
		time.Sleep(config.StopTime)
		// Kill 0 does nothing but still checks for errors
		//  if there is no error, the process is still running
		if err = syscall.Kill(pid, 0); err == nil {
			slog.Warn("program did not exit cleanly, killing...",
				slog.String("program", config.name),
				slog.String("Time to wait", config.StopTime.String()),
				slog.String("signal", stopSignal.String()),
				slog.Int("pid", pid),
			)
			err = syscall.Kill(pid, syscall.SYS_KILL)
			if err != nil {
				slog.Error("could not kill program",
					slog.String("program", config.name),
					slog.String("error", err.Error()),
					slog.Int("pid", pid),
				)
			}
		}
	}
	return nil
}
