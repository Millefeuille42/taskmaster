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

// TODO move regexes into a utils file

const config_regex_str = "^\\s*(?:;|#).*$|^\\s*\\[\\s*([\\w _-]+?)\\s*\\]\\s*?(?:(?:;|#).*)?$|^\\s*([^=#;\\s]+)\\s*=\\s*(?:\"([^\"]*)\"|'([^']*)'|([^;\\s=#]+))\\s*?(?:(?:;|#).*)?$";
const section_regex_str = "^[\r\n\t\f\v ]*\\[[\r\n\t\f\v ]*([a-zA-Z0-9 _-]+)[\r\n\t\f\v ]*\\][\r\n\t\f\v ]*$";

const kv_regex_str = "^[\r\n\t\f\v ]*([^=#;\r\n\t\f\v ]+)[\r\n\t\f\v ]*=([^=\"'#;\r\n\t\f\v ]+)[\r\n\t\f\v ]*$";
const kv_regex_dq_str = "^[\r\n\t\f\v ]*([^=#;\r\n\t\f\v ]+)[\r\n\t\f\v ]*=(\"[^\"]*\")[\r\n\t\f\v ]*$";
const kv_regex_sq_str = "^[\r\n\t\f\v ]*([^=#;\r\n\t\f\v ]+)[\r\n\t\f\v ]*=('[^']*')[\r\n\t\f\v ]*$";

fn compile_regex(re_str: str) regex::regex = {
    match(regex::compile(re_str)) {
        case let err: regex::error =>
            log::lfatalf(log::global, "Error: could not compile regex {}: {}", re_str, regex::strerror(err));
        case let re: regex::regex => return re;
    };
};

fn get_regex(re: *(regex::regex|void), re_str: const *str) regex::regex = {
    *re = match (*re) {
        case let re: regex::regex => yield re;
        case void => yield compile_regex(*re_str);
    };
    return *re as regex::regex;
};

fn get_kv_regexes() [3](str, regex::regex) = {
    static let kv_regex: (regex::regex|void) = void;
    static let kv_regex_dq: (regex::regex|void) = void;
    static let kv_regex_sq: (regex::regex|void) = void;

    return [
        ("kv_regex_dq_str", get_regex(&kv_regex_dq, &kv_regex_dq_str)),
        ("kv_regex_sq_str", get_regex(&kv_regex_sq, &kv_regex_sq_str)),
        ("kv_regex_str", get_regex(&kv_regex, &kv_regex_str))
    ];
};

fn parse_kv(line: *str, config: *hfo_types::raw_program_config) void = {
    const kv_regexes = get_kv_regexes();

    for (let kv_re .. kv_regexes) {
        if (regex::test(&kv_re.1, *line)) {
            const result = regex::find(&kv_re.1, *line);
            defer regex::result_free(result);
            if (len(result) < 3) {
                log::lprintfln(log::global, "Error while parsing line: {}", line);
                continue;
            };
            // log::lprintfln(log::global, "'{}' matched {}", *line, kv_re.0);
            let key: str = result[1].content;
            let val: str = strings::trim(result[2].content, '"', '\'');
            // log::lprintfln(log::global,"{}: {}", key, val);
            map_kv_to_raw(&key, &val, config);
            return;
        };
    };

    log::lprintfln(log::global, "invalid line: {}, ignoring...", *line);
    return;
};

fn parse_file(path: str) ([]hfo_types::program_config | fs::error) = {
    log::lprintfln(log::global, "reading: {}", path);

    static let s_section_re: (regex::regex|void) = void;
    const section_re = get_regex(&s_section_re, &section_regex_str);

    const file = os::open(path)?;
    defer io::close(file)!;

    let configs: []hfo_types::program_config = [];
    let section: str = "";
    let raw_config = hfo_types::new_raw_program_config();
    const scan = bufio::newscanner(file, types::SIZE_MAX);
    for (let i = 1u; true; i += 1) {
        const line = match (bufio::scan_line(&scan)?) {
            case io::EOF => break;
            case let line: const str => yield line;
        };
        if (regex::test(&section_re, line)) {
            const result = regex::find(&section_re, line);
            defer regex::result_free(result);
            if (len(result) < 2) {
                log::lprintfln(log::global, "Error while parsing section: {}", line);
                continue;
            };
            let capture = result[1].content;
            log::lprintfln(log::global, "found section: {}", capture);
            if (section != "") {
                log::lprintfln(log::global, "commiting section: {}", section);
                append(configs, raw_config_to_config(&raw_config));
            };
            log::lprintfln(log::global, "starting new section");
            raw_config = hfo_types::new_raw_program_config();
            section = capture;
            raw_config.name = section;
        } else { parse_kv(&line, &raw_config); };
    }; 
    log::lprintfln(log::global, "commiting section: {}", section);
    append(configs, raw_config_to_config(&raw_config));
    return configs;
};

export fn load_config(path: str) ([]hfo_types::program_config | fs::error) = {
    const path = os::realpath(path)?;
    const stat = os::stat(path)?;
    let configs: []hfo_types::program_config = [];

    if (!fs::isdir(stat.mode)) {
        return parse_file(path);
    };
    
    const path_buf = path::init(path)!;
    let iter = os::iter(path)?;
    defer fs::finish(iter); 
    for (const entry => fs::next(iter)!) {
        if (
            entry.name == "." 
            || entry.name == ".."
            || !strings::hassuffix(entry.name, ".ini")
        ) 
        {continue;};
        const entry_path = path::push(&path_buf, entry.name)!;
        defer path::pop(&path_buf);
        const entry_path = os::realpath(entry_path)?;
        const stat = os::stat(entry_path)?;
        if (fs::isdir(stat.mode)) {continue;};
        match (parse_file(path::string(&path_buf))) {
            case let err: fs::error => 
                log::lprintfln(log::global, "Error: could not parse {}: {}", entry_path, fs::strerror(err));
            case let confs: []hfo_types::program_config => 
                append(configs, confs...);
        };
    };

    return configs;
};
