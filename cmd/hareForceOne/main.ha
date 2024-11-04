use os;
use os::exec;
use fmt;
use time;
use unix::signal;
use path;
use fs;
use io;
use errors;

fn fill_test_struct() program_config = {
	let config = program_config {
		name = "echo",
		command = exec::cmd("echo", "Hello!")!,
		numprocs = 3,
		autostart = true,
		restart_when = restart_when::NEVER,
		exit_codes = [1, 2],
		start_time = 2 * time::SECOND,
		stop_time = 2 * time::SECOND,
		max_restarts = 2,
		stop_signal = signal::sig::TERM,
		stdout = path::init("./log.txt")!,
		stderr =  path::init("/dev/stderr")!,
		umask = fs::mode::USER_RWX | fs::mode::GROUP_RWX | fs::mode::OTHER_RWX,
	};
	return config;
};

fn start_program(config: *program_config) program_info = {
	let process = exec::start(&config.command : *exec::command)!;
	let program = program_info {
		config = config,
		process = process,
		restart = 0,
	};
	return program;
};

fn process_test() void = {
	let config1 = fill_test_struct();
	let config2 = fill_test_struct();
	let config3 = fill_test_struct();
	let program1 = start_program(&config1);
	let program2 = start_program(&config2);
	let program3 = start_program(&config3);

	let process4 = match (exec::fork()) {
	case let process: exec::process =>
		yield process;
	case => {
		time::sleep(time::SECOND * 6);
		fmt::printfln("Bye!")!;
		os::exit(42);
	};
	case let error: exec::error =>
	   fmt::fatalf("Error forking");
	};

	fmt::printfln("Program 1 {}", program1.process: uint)!;
	fmt::printfln("Program 2 {}", program2.process: uint)!;
	fmt::printfln("Program 3 {}", program3.process: uint)!;
	fmt::printfln("Process 4 {}", process4: int)!;

	let result = exec::waitany()!;
	fmt::printfln("Process {} just finished with status code: {}", result.0: uint, exec::exit(&result.1): exec::exited: int )!;

	time::sleep(time::SECOND * 1);

	fmt::printfln("Parent process is tired, taking a nap...")!;
	time::sleep(time::SECOND * 4);
	fmt::printfln("Parent process is still tired but is back to work...")!;

	let result = exec::waitany()!;
	fmt::printfln("Process {} just finished with status code: {}", result.0: uint, exec::exit(&result.1): exec::exited: int )!;

	let result = exec::waitany()!;
	fmt::printfln("Process {} just finished with status code: {}", result.0: uint, exec::exit(&result.1): exec::exited: int )!;

	let result = exec::waitany()!;
	fmt::printfln("Process {} just finished with status code: {}", result.0: uint, exec::exit(&result.1): exec::exited: int )!;

	// should crash
	// let result = exec::waitany()!;
	// fmt::printfln("Process {} just finished with status code: {}", result.0: uint, exec::exit(&result.1): exec::exited: int )!;

	match (exec::waitany()) {
	case =>
		fmt::printfln("Curious mh 🤔")!;
	case exec::error =>
		fmt::printfln("No child left to wait")!;
	};
};

fn shared_memory_test() void = {
	// create global shared memory
	// gives a file handled that can be used to lock
	// let file = os::shm_open("shared", fs::flag::RDWR,  fs::mode::USER_RWX | fs::mode::GROUP_RWX | fs::mode::OTHER_RWX)!;
	// io::trunc(file, 4096)!;
	// let ptr = match(io::mmap(null, 4096, io::prot::READ | io::prot::WRITE, io::mflag::SHARED, file, 0)) {
	// case let ptr: *opaque =>
	// 	yield ptr;
	// case errors::error =>
	// 	fmt::fatalf("Error mapping memory");
	// };

	// let ptr_test = ptr: *[4096]u8;

	// ptr_test[0] = 0;
	// ptr_test[1] = 1;
	// ptr_test[2] = 2;
	// ptr_test[3] = 3;

	// fmt::printfln("print item from slice {}", ptr_test[1])!;

	match(exec::fork()) {
	case let process: exec::process =>
		let file = os::shm_open("shared", fs::flag::RDWR | fs::flag::CREATE, fs::mode::USER_RWX | fs::mode::GROUP_RWX | fs::mode::OTHER_RWX)!;
		io::trunc(file, 4096)!;

		let ptr = match(io::mmap(null, 4096, io::prot::READ | io::prot::WRITE, io::mflag::SHARED, file, 0)) {
		case let ptr: *opaque =>
			yield ptr;
		case errors::error =>
			fmt::fatalf("Error mapping memory");
		};

		let ptr_test = ptr: *[4096]u8;

		fmt::printfln("Parent locking!")!;
		io::lock(file, true, io::lockop::EXCLUSIVE)!;
		fmt::printfln("Parent locked!")!;

		time::sleep(time::SECOND * 2);
		ptr_test[1] = 42;
		fmt::printfln("print item from slice {}", ptr_test[1])!;

		fmt::printfln("Parent unlocking!")!;
		io::lock(file, true, io::lockop::UNLOCK)!;
		exec::wait(&process)!;
		io::munmap(ptr, 4096)!;

	case void =>
		time::sleep(time::SECOND / 2);
		let file = os::shm_open("shared", fs::flag::RDWR,  fs::mode::USER_RWX | fs::mode::GROUP_RWX | fs::mode::OTHER_RWX)!;
		// io::trunc(file, 4096)!;

		let ptr = match(io::mmap(null, 4096, io::prot::READ | io::prot::WRITE, io::mflag::SHARED, file, 0)) {
		case let ptr: *opaque =>
			yield ptr;
		case errors::error =>
			fmt::fatalf("Error mapping memory");
		};

		let ptr_test = ptr: *[4096]u8;

		fmt::printfln("Child locking!")!;
		io::lock(file, true, io::lockop::EXCLUSIVE)!;
		fmt::printfln("Child locked!")!;

		time::sleep(time::SECOND * 2);
		fmt::printfln("print item from slice {}", ptr_test[1])!;
		ptr_test[1] = 2;

		fmt::printfln("Child unlocking!")!;
		io::lock(file, true, io::lockop::UNLOCK)!;

		io::munmap(ptr, 4096)!;
		// os::shm_unlink("shared")!;
		os::exit(1);
	case let error: exec::error =>
	   fmt::fatalf("Error forking");
	};

	// fmt::printfln("print item from slice {}", ptr_test[1])!;
	// io::munmap(ptr, 4096)!;
	os::shm_unlink("shared")!;
};

export fn main() void = {
	if (len(os::args) <= 1) {
		fmt::printf("usage: {} <path to a config file or folder>\n", os::args[0])!;
		return;
	};

	// shared_memory_test();
	return;
};