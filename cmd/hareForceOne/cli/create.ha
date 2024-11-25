use ascii;
use bufio;
use encoding::utf8;
use errors;
use fmt;
use fs;
use io;
use log;
use os;
use os::exec;
use strings;

const PROMPT = "hareForceOne $> ";

type error = !(io::error | utf8::invalid);

type command_type = enum {
	NONE = 0,
	INVALID,
	STATUS,
	RELOAD,
	START,
	STOP,
	RESTART,
	HISTORY,
	EXIT,
	HELP,
};

type command = struct {
	cmd_type: command_type,
	args: []str,
};

// read line
// parse input
// send to socket
// wait for answer
// display infod

fn read_line() (str | io::EOF | error) = {
	fmt::printf(PROMPT)?;
	bufio::flush(os::stdout)?;
	return match (bufio::read_line(os::stdin)?) {
	case let line: []u8 =>
		yield strings::fromutf8(line)?;
	case let eof: io::EOF =>
		yield eof;
	};
};

fn split(line: str) []str = strings::split(line, " \t");

fn to_lower(word: str) str = ascii::strlower_buf(word, strings::toutf8(word));

fn build_cmd(words: []str) command = {
	let cmd = command { cmd_type = command_type::NONE, ...};
	let cpy_name = strings::dup(words[0]);
	defer free(cpy_name);

	let cmd_type = switch(to_lower(cpy_name)) {
	case "status" => yield command_type::STATUS;
	case "reload" => yield command_type::RELOAD;
	case "start" => yield command_type::START;
	case "stop" => yield command_type::STOP;
	case "restart" => yield command_type::RESTART;
	case "history" => yield command_type::HISTORY;
	case "help" => yield command_type::HELP;
	case =>
		log::lprintfln(log::global,
			"\"{}\" is not a recognized command", words[0]);
		yield command_type::INVALID;
	};

	fmt::printfln("cmd: [{}], type: [{}]", words[0], cmd_type: int)!;
	cmd.cmd_type = cmd_type;

	cmd.args = words[1..];
	
	return cmd;
};

fn cli_main() void = {
	for (true) {
		let words = match(read_line()) {
		case let line: str =>
			yield split(line);
		case io::EOF =>
			fmt::printfln("\n[EOF]")!;
			break ;
		case error =>
			log::lfatalf(log::global, "Error: in cli_main()");
		};
		defer free(words);

		let command = build_cmd(words);

		for (let arg .. command.args) {
			fmt::printfln("[{}]", arg)!;
			// let word_1 = ascii::strupper(word);
			// defer free(word_1);
			// fmt::printfln("[{}]", word_1)!;
			// let word_2 = ascii::strupper_buf(word, strings::toutf8(word));
			// fmt::printfln("[{}]", word_2)!;
		};
	};
};

export fn create() exec::process = {
	let cli_process = match(exec::fork()) {
	case let process: exec::process =>
		exec::wait(&process)!;
		yield process;
	case void =>
		cli_main();
		os::exit(0);
	case let err: exec::error =>
		log::lfatalf(log::global, "Error: {}", exec::strerror(err));
	};

	return cli_process;
};