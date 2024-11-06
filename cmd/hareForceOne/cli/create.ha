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

type error = !(io::error | utf8::invalid);

const PROMPT = "hareForceOne $> ";

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

// fn parse_line()

fn cli_main() void = {
	fmt::println("hello")!;
	match(read_line()!) {
	case let line: str =>
		fmt::printfln("line is: [{}]", line)!;
	case io::EOF =>
		fmt::printfln("EOF")!;
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