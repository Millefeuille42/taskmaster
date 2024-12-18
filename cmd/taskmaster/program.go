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

func runProgram(config *Config) error {
	// TODO make it so it uses a "running program" store instead of config
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

func stopProgram(config *Config) error {
	// TODO make it so it uses a "running program" store instead of config
	if config.pid == 0 {
		slog.Error(
			"Could not stop program, invalid PID",
			slog.String("program", config.name),
			slog.Int("pid", config.pid),
		)
		return nil
	}

	var stopSignal syscall.Signal
	switch config.StopSignal {
	case "ABRT":
		stopSignal = syscall.SIGABRT
	case "ALRM":
		stopSignal = syscall.SIGALRM
	case "BUS":
		stopSignal = syscall.SIGBUS
	case "CHLD":
		stopSignal = syscall.SIGCHLD
	case "CONT":
		stopSignal = syscall.SIGCONT
	case "FPE":
		stopSignal = syscall.SIGFPE
	case "HUP":
		stopSignal = syscall.SIGHUP
	case "ILL":
		stopSignal = syscall.SIGILL
	case "INT":
		stopSignal = syscall.SIGINT
	case "KILL":
		stopSignal = syscall.SIGKILL
	case "PIPE":
		stopSignal = syscall.SIGPIPE
	case "QUIT":
		stopSignal = syscall.SIGQUIT
	case "SEGV":
		stopSignal = syscall.SIGSEGV
	case "STOP":
		stopSignal = syscall.SIGSTOP
	case "TERM":
		stopSignal = syscall.SIGTERM
	case "TSTP":
		stopSignal = syscall.SIGTSTP
	case "TTIN":
		stopSignal = syscall.SIGTTIN
	case "TTOU":
		stopSignal = syscall.SIGTTOU
	case "USR1":
		stopSignal = syscall.SIGUSR1
	case "USR2":
		stopSignal = syscall.SIGUSR2
	case "POLL":
		stopSignal = syscall.SIGPOLL
	case "PROF":
		stopSignal = syscall.SIGPROF
	case "SYS":
		stopSignal = syscall.SIGSYS
	case "TRAP":
		stopSignal = syscall.SIGTRAP
	case "URG":
		stopSignal = syscall.SIGURG
	case "VTALRM":
		stopSignal = syscall.SIGVTALRM
	case "XCPU":
		stopSignal = syscall.SIGXCPU
	case "XFSZ":
		stopSignal = syscall.SIGXFSZ
	default:
		return errors.New("invalid or unsupported signal: " + config.StopSignal)
	}

	slog.Debug(
		"stop",
		slog.String("program", config.name),
		slog.String("signal", stopSignal.String()),
		slog.Int("pid", config.pid),
	)
	return syscall.Kill(config.pid, stopSignal)
}
