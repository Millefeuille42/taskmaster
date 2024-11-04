export type restart_when = enum {
    NEVER,
    UNEXPECTED,
    ALWAYS,
};

export fn restart_when_to_str(src: *restart_when) str =  {
    switch (*src) {
        case restart_when::NEVER => return "NEVER";
        case restart_when::UNEXPECTED => return "UNEXPECTED";
        case restart_when::ALWAYS => return "ALWAYS";
        case => return "INVALID";
    };
};