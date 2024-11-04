export type raw_program_config = struct {
    name: str,
    command: str,
    numprocs: str,
    autostart: str,
    restart_when: str,
    exit_codes: str,
    start_time: str,
    stop_time: str,
    max_restarts: str,
    stop_signal: str,
    stdout: str,
    stderr: str,
    umask: str,
    env: str,
    workdir: str
};

export fn new_raw_program_config() raw_program_config = {
    return raw_program_config {
        name = "",
        command = "",
        numprocs = "",
        autostart = "",
        restart_when = "",
        exit_codes = "",
        start_time = "",
        stop_time = "",
        max_restarts = "",
        stop_signal = "",
        stdout = "",
        stderr = "",
        umask = "",
        env = "",
        workdir = ""
    };
};
