package main

import (
	"errors"
	"fmt"
	"syscall"
)

func printProgramStatus(config Config) {
	// TODO Handle multiple procs
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

func stringToSignal(strSignal string) (syscall.Signal, error) {
	var signal syscall.Signal

	switch strSignal {
	case "ABRT":
		signal = syscall.SIGABRT
	case "ALRM":
		signal = syscall.SIGALRM
	case "BUS":
		signal = syscall.SIGBUS
	case "CHLD":
		signal = syscall.SIGCHLD
	case "CONT":
		signal = syscall.SIGCONT
	case "FPE":
		signal = syscall.SIGFPE
	case "HUP":
		signal = syscall.SIGHUP
	case "ILL":
		signal = syscall.SIGILL
	case "INT":
		signal = syscall.SIGINT
	case "KILL":
		signal = syscall.SIGKILL
	case "PIPE":
		signal = syscall.SIGPIPE
	case "QUIT":
		signal = syscall.SIGQUIT
	case "SEGV":
		signal = syscall.SIGSEGV
	case "STOP":
		signal = syscall.SIGSTOP
	case "TERM":
		signal = syscall.SIGTERM
	case "TSTP":
		signal = syscall.SIGTSTP
	case "TTIN":
		signal = syscall.SIGTTIN
	case "TTOU":
		signal = syscall.SIGTTOU
	case "USR1":
		signal = syscall.SIGUSR1
	case "USR2":
		signal = syscall.SIGUSR2
	case "POLL":
		signal = syscall.SIGPOLL
	case "PROF":
		signal = syscall.SIGPROF
	case "SYS":
		signal = syscall.SIGSYS
	case "TRAP":
		signal = syscall.SIGTRAP
	case "URG":
		signal = syscall.SIGURG
	case "VTALRM":
		signal = syscall.SIGVTALRM
	case "XCPU":
		signal = syscall.SIGXCPU
	case "XFSZ":
		signal = syscall.SIGXFSZ
	default:
		return signal, errors.New("invalid or unsupported signal: " + strSignal)
	}

	return signal, nil
}
