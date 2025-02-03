package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func isExitCodeValid(config Config, status ProgramStatus) bool {
	for _, code := range config.ExitCodes {
		if code == status.ExitCode {
			return true
		}
	}
	return false
}

func countRunningPids(config Config) int {
	count := 0
	for _, pid := range config.pids {
		if pid.Running {
			count++
		}
	}
	return count
}

func startProc(config Config, configChannel chan<- Config) {
	for i := countRunningPids(config); i < config.NumProcs; i++ {
		go func() {
			err := runProgram(config, configChannel)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
	}
}

func printProgramStatus(config Config) {
	fmt.Printf("%s %s", config.name, config.Command)
	if config.NumProcs > 1 {
		fmt.Printf(" (%d)", config.NumProcs)
	}
	fmt.Println(":")
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
		// Per spec, SIGPOLL is the same number as SIGIO
		//  Doing this since syscall.SIGPOLL doesn't exist on macOS
		signal = syscall.SIGIO
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
