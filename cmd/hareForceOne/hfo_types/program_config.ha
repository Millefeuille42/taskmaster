use os;
use fmt;
use log;
use fs;
use path;
use regex;
use types;
use bufio;
use io;
use strings;
use strconv;
use time;
use unix::signal;
use os::exec;

export type program_config = struct {
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

export fn print_program_config(src: *program_config) void = {
    fmt::printfln("name: \"{}\"", src.name)!;
    // TODO fmt::printfln("command: \"{}\"", src.command)!;
    fmt::printfln("numprocs: \"{}\"", src.numprocs)!;
    fmt::printfln("autostart: \"{}\"", src.autostart)!;
    fmt::printfln("restart_when: \"{}\"", restart_when_to_str(&src.restart_when))!;
    fmt::println("exit_codes: ")!;
    for (let code .. src.exit_codes) {
        fmt::printfln("\t{}", code)!;
    };
    fmt::printfln("start_time: \"{}\"", src.start_time)!;
    fmt::printfln("stop_time: \"{}\"", src.stop_time)!;
    fmt::printfln("max_restarts: \"{}\"", src.max_restarts)!;
    fmt::printfln("stop_signal: \"{}\"", src.stop_signal: int)!;
    // TODO fmt::printfln("stdout: \"{}\"", src.stdout)!;
    // TODO fmt::printfln("stderr: \"{}\"", src.stderr)!;
    fmt::printfln("umask: \"{}\"", src.umask: int)!;
};

export fn new_program_config() program_config = {
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
