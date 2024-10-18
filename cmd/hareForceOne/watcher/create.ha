use os::exec;

fn init_shared_memory() *opaque = {

};

export fn create() exec::process = {
	let process = match(exec::fork()) {
	case let process: exec::process =>

	case void =>

	case let err: exec::error =>
		log::lfatalf(log::global, "Error: {}", fs::strerror(err));
	};
};