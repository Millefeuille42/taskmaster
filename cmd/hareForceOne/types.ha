use fs;
use io;
use os;
use os::exec;
use path;
use time;
use unix::signal;

type restart_when = enum {
    NEVER,
    UNEXPECTED,
    ALWAYS,
};

type program_config = struct {
    name: str,
    command: (exec::command | void),
    numprocs: uint,
    autostart: bool,
    restart_when: restart_when,
    exit_codes: []uint,
    start_time: time::duration,
    stop_time: time::duration,
    max_restarts: uint,
    stop_signal: signal::sig,
    stdout: (path::buffer | io::file | exec::nullfd),
    stderr: (path::buffer | io::file | exec::nullfd),
    umask: fs::mode,
};

fn new_program_config() program_config = {
    return program_config {
        name = "",
        command = void,
        numprocs = 0,
        autostart = false,
        restart_when = restart_when::NEVER,
        exit_codes = [],
        start_time = 0,
        stop_time = 0,
        max_restarts = 0,
        stop_signal = signal::sig::TERM,
        stdout = os::stdout_file,
        stderr = os::stderr_file,
        umask = fs::mode::UNKNOWN,
    };
};