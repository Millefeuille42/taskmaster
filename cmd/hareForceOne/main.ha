use os;
use fmt;

export fn main() void = {
    if (len(os::args) <= 1) {
        fmt::printf("usage: {} <path to a config file or folder>\n", os::args[0])!;
        return;
    };

    return;
};