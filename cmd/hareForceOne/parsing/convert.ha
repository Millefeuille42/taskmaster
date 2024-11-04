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
use hfo_types;

fn raw_config_to_restart(val: *str) hfo_types::restart_when = {
    switch (*val) {
        case "never" => return hfo_types::restart_when::NEVER;
        case "unexpected" => return hfo_types::restart_when::UNEXPECTED;
        case "always" => return hfo_types::restart_when::ALWAYS;
        case => log::lprintfln(log::global, "Error: unable to process restart_when value: {}", *val); return hfo_types::restart_when::NEVER;
    };
};

fn raw_config_to_signal(val: *str) signal::sig = {
    switch (*val) {
        case "HUP" => return signal::sig::HUP;
        case "INT" => return signal::sig::INT;
        case "QUIT" => return signal::sig::QUIT;
        case "ILL" => return signal::sig::ILL;
        case "TRAP" => return signal::sig::TRAP;
        case "ABRT" => return signal::sig::ABRT;
        case "BUS" => return signal::sig::BUS;
        case "FPE" => return signal::sig::FPE;
        case "KILL" => return signal::sig::KILL;
        case "USR1" => return signal::sig::USR1;
        case "SEGV" => return signal::sig::SEGV;
        case "USR2" => return signal::sig::USR2;
        case "PIPE" => return signal::sig::PIPE;
        case "ALRM" => return signal::sig::ALRM;
        case "TERM" => return signal::sig::TERM;
        case "CHLD" => return signal::sig::CHLD;
        case "CONT" => return signal::sig::CONT;
        case "STOP" => return signal::sig::STOP;
        case "TSTP" => return signal::sig::TSTP;
        case "TTIN" => return signal::sig::TTIN;
        case "TTOU" => return signal::sig::TTOU;
        case "URG" => return signal::sig::URG;
        case "XCPU" => return signal::sig::XCPU;
        case "XFSZ" => return signal::sig::XFSZ;
        case "VTALRM" => return signal::sig::VTALRM;
        case "PROF" => return signal::sig::PROF;
        case "WINCH" => return signal::sig::WINCH;
        case "IO" => return signal::sig::IO;
        case "POLL" => return signal::sig::POLL;
        case "PWR" => return signal::sig::PWR;
        case "SYS" => return signal::sig::SYS;
        case => log::lprintfln(log::global, "Error: unable to process stop_signal value: {}", *val); return signal::sig::INT;
    };
};

fn raw_config_to_uint(val: *str) uint = {
    return match (strconv::stou(*val)) {
        case let val_as_num: uint => yield val_as_num;
        case => log::lprintfln(log::global, "Error: unable to process number: {}", *val); yield 0;
    };
};

fn raw_config_to_duration(val: *str) time::duration = {
    static const units: [5](str, time::duration) = [
        ("ns", time::NANOSECOND),
        ("us", time::MICROSECOND),
        ("ms", time::MILLISECOND),
        ("s", time::SECOND),
        ("m", time::MINUTE)
    ];

    for (let unit .. units) {
        if (strings::hassuffix(*val, unit.0)) {
            return raw_config_to_uint(&strings::trimsuffix(*val, unit.0)): int * unit.1;
        };
    };
    log::lprintfln(log::global, "Error: unable to process duration: {}", *val);
    return 0;
};

fn raw_config_to_uint_array(val: *str, dest: *[]uint) void = {
    let vals: []str = strings::split(*val, ",");
    defer free(vals);
    for (let val .. vals) {
        append(dest, raw_config_to_uint(&val));
    };
};

fn raw_config_to_fd(val: *str, default: io::file) (path::buffer | io::file | exec::nullfd) = {
    switch (*val) {
        case "discard" => return exec::nullfd;
        case "" => return default;
        case => return path::init(*val)!;
    };
};

fn raw_config_to_cmd(command: *str, workdir: *str, env: *str, config: *hfo_types::program_config) (exec::command | void) = {
    let argv: []str = strings::split(*command, " ");
    defer free(argv);

    let cmd = match(exec::cmd(argv[0], strings::dupall(argv[1..])...)) {
        case let err: exec::error => log::lprintfln(log::global, "Error: unable to process command: {}", *command); return;
        case let cmd: exec::command => yield cmd;
    };
    exec::chdir(&cmd, strings::dup(*workdir));
    cmd.env = strings::split(*env, ",");

    return cmd;
};

fn map_kv_to_raw(key: *str, val: *str, config: *hfo_types::raw_program_config) void = {
    switch (*key) {
        case "command" => config.command = *val;
        case "numprocs" => config.numprocs = *val; 
        case "autostart" => config.autostart = *val; 
        case "restart_when" => config.restart_when = *val; 
        case "exit_codes" => config.exit_codes = *val; 
        case "start_time" => config.start_time = *val; 
        case "stop_time" => config.stop_time = *val; 
        case "max_restarts" => config.max_restarts = *val; 
        case "stop_signal" => config.stop_signal = *val; 
        case "stdout" => config.stdout = *val; 
        case "stderr" => config.stderr = *val; 
        case "env" => config.env = *val; 
        case "workdir" => config.workdir = *val; 
        case "umask" => config.umask = *val;
        case => log::lprintfln(log::global, "Error: unknown key \"{}\"", *key);
    };
};

fn raw_config_to_config(raw_config: *hfo_types::raw_program_config) hfo_types::program_config = {
    let config = hfo_types::new_program_config();
    config.name = raw_config.name;
    config.numprocs = raw_config_to_uint(&raw_config.numprocs);
    config.max_restarts = raw_config_to_uint(&raw_config.max_restarts);
    config.umask = raw_config_to_uint(&raw_config.umask): fs::mode;
    config.autostart = raw_config.autostart == "true";
    config.start_time = raw_config_to_duration(&raw_config.start_time);
    config.stop_time = raw_config_to_duration(&raw_config.stop_time);
    config.restart_when = raw_config_to_restart(&raw_config.restart_when);
    config.stop_signal = raw_config_to_signal(&raw_config.stop_signal);
    config.stdout = raw_config_to_fd(&raw_config.stdout, os::stdout_file);
    config.stderr = raw_config_to_fd(&raw_config.stderr, os::stderr_file);
    raw_config_to_uint_array(&raw_config.exit_codes, &config.exit_codes);
    return config;
};