use os;
use fmt;

use cli;

export fn main() void = {
    if (len(os::args) <= 1) {
        fmt::printf("usage: {} <path to a config file or folder>\n", os::args[0])!;
        return;
    };

	let cli_process = cli::create();
	// watcher::create();
	// shared memory setup
	// parse config
	// store config in shared memory setup
	// open unix socket

	// listen socket
	// parse exec stuff from cli
	// format info
	// send info back to cli

    return;
};