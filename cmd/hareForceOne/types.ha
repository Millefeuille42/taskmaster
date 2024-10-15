use fs;
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
    command: exec::command,
    numprocs: uint,
    autostart: bool,
    restart_when: restart_when,
    exit_codes: []uint,
    start_time: time::duration,
    stop_time: time::duration,
    max_restarts: uint,
    stop_signal: signal::sig,
    stdout: (path::buffer | exec::nullfd),
    stderr: (path::buffer | exec::nullfd),
    umask: fs::mode,
};