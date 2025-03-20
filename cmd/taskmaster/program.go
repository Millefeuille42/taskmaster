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

func processConfigDiff(config Config, oldConfig Config, configChannel chan<- Config) {
	if config.HasCriticalChange(oldConfig) {
		if err := stopProgram(oldConfig, configChannel); err != nil {
			slog.Error("An error occurred while stopping program",
				slog.String("error", err.Error()),
				slog.String("program", config.name),
			)
		}
	}
	startProc(config, configChannel)
}

func handleReload(
	configs map[string]Config,
	configChannel chan<- Config,
) map[string]Config {
	oldConfigs := configs
	configs = parseConfig()

	for name, config := range oldConfigs {
		// Stop programs that are not featured in the new config
		if configs[name].name == "" {
			if err := stopProgram(config, configChannel); err != nil {
				slog.Error("An error occurred while stopping program",
					slog.String("error", err.Error()),
					slog.String("program", config.name),
				)
			}
		}
	}

	for name, config := range configs {
		// Start new autostart programs
		if oldConfigs[name].name == "" && config.AutoStart {
			startProc(config, configChannel)
			continue
		}
		// If the program is already featured in the config
		//  check if any critical element has changed and restart it if so
		config.lock.Lock()
		config.pids = clonePidsMap(oldConfigs[name].pids)
		config.lock.Unlock()
		processConfigDiff(config, oldConfigs[name], configChannel)
		configs[name] = config
	}

	return configs
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
			configs = handleReload(configs, configChannel)
			slog.Info("Reloaded configuration")
		case args := <-command:
			if cmd, ok := commands[args[0]]; ok {
				err = cmd.Function(statusCommand, configChannel, &configs, args)
				if err != nil {
					slog.Error(err.Error(), slog.String("command", args[0]))
				}
				continue
			}
			statusCommand <- "Unknown command: " + args[0]
		case config := <-configChannel:
			pids := config.pids
			config.lock.Lock()
			for pid, status := range pids {
				if status.Running != false {
					continue
				}
				delete(config.pids, pid)
				if status.ExitCode == -1 {
					// This means it has been stopped by a signal
					//  thus it is highly possible that it has been shutdown
					//  with the stop command
					stopSignal, _ := stringToSignal(config.StopSignal)
					if int(status.SysStatus) == int(stopSignal) {
						slog.Warn("Program has been stopped by TUI",
							slog.String("program", config.name),
						)
						continue
					}
					slog.Warn("Program has been stopped by an invalid signal",
						slog.Int("signal", int(stopSignal)),
						slog.Int("expected signal", int(status.SysStatus)),
						slog.String("program", config.name),
					)
				}
				if config.restart >= config.MaxRestarts {
					slog.Warn("Program has reached max number of restarts",
						slog.Int("maxRestarts", config.MaxRestarts),
						slog.Int("restarts", config.restart),
						slog.String("program", config.name),
					)
					continue
				} else {
					config.restart += 1
				}
				if status.ExitedEarly || !isExitCodeValid(config, status) {
					slog.Warn("Program has stopped unexpectedly",
						slog.Int("", status.ExitCode),
						slog.String("program", config.name),
					)
					if config.RestartWhen == "unexpected" {
						slog.Info("Restarting program",
							slog.String("program", config.name),
							slog.Int("restarts", config.restart),
						)
						config.lock.Unlock()
						startProc(config, configChannel)
						config.lock.Lock()
						continue
					}
				}
				if config.RestartWhen == "always" {
					slog.Info("Restarting program",
						slog.String("program", config.name),
						slog.Int("restarts", config.restart),
					)
					config.lock.Unlock()
					startProc(config, configChannel)
					config.lock.Lock()
				}
			}
			config.lock.Unlock()
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
	config.lock.Lock()
	config.pids[pid] = ProgramStatus{
		Running:     true,
		ExitedEarly: false,
		ExitCode:    0,
	}
	config.lock.Unlock()
	configChannel <- config

	startTime := time.Now().Add(config.StartTime)
	slog.Info("Starting program",
		slog.String("program", config.name),
		slog.String("command", strings.Join(cmd.Args, " ")),
		slog.Int("pid", pid),
	)
	err = cmd.Wait()
	status := ProgramStatus{
		Running:     false,
		ExitedEarly: time.Now().Before(startTime),
		ExitCode:    cmd.ProcessState.ExitCode(),
	}
	if waitStatus, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
		status.SysStatus = waitStatus
	}
	config.lock.Lock()
	config.pids[pid] = status
	config.lock.Unlock()
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
		slog.Info(
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
