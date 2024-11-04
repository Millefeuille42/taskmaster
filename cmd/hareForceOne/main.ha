use os;
use fmt;
use log;
use fs;
use parsing;
use hfo_types;

export fn main() void = {
    if (len(os::args) <= 1) {
        fmt::printf("usage: {} <path to a config file or folder>\n", os::args[0])!;
        return;
    };

    const configs = match(parsing::load_config(os::args[1])) {
        case let err: fs::error => 
            log::lfatalf(log::global, "Error: {}", fs::strerror(err));
        case let configs: []hfo_types::program_config =>
            yield configs;
    };
    for (let config .. configs) {
        hfo_types::print_program_config(&config);
    };
    return;
};