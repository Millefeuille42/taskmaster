// YES I KNOW ITS STUPID
use fmt;

export type color = enum {
	RED,
	BLACK,
};

export type data = (int, str);
export type compare_function = fn(data, data) uint;
export type allocate_function = fn(node) *opaque;

export type node = struct {
	color: color,
	left: *node,
	right: *node,
	parent: *node,
	data: data,
};

export type tree = struct {
	root: nullable *node,
	nil: node,
	node_count: uint,
	compare: *compare_function,
	allocate: nullable *allocate_function,
};

export fn init(tree: *tree) void = {
	tree.root = &(tree.nil);
	tree.node_count = 0;
	tree.nil.color = color::BLACK;
	// tree.compare = 
	assert(tree.compare != null);
};

fn create_node(tree: *tree, data: data) *node = {
	let new_node: node = node {
		color = color::RED,
		left = &tree.nil,
		right = &tree.nil,
		parent = &tree.nil,
		data = data,
	};

	let node: *node = match(tree.allocate) {
	case null =>
		yield alloc(new_node);
	case let ptr: *allocate_function =>
		yield ptr(new_node): *node;
	};
	return node;
};


export fn add(tree: *tree, data: data) void = {
	let node: *node = create_node(tree, data);

	insert_node(tree, node);
};

fn insert_node(tree: *tree, n: *node) void = {
	let root: *node = match(tree.root) {
	case null =>
		tree.root = n;
		tree.node_count += 1;
		return;
	case let n: *node =>
		yield n;
	};

	let x = root;
	let y = &tree.nil;

	for (x != &tree.nil) {
		y = x;
		if (tree.compare(n.data, x.data) < 0) {
			x = x.left;
		} else {
			x = x.right;
		};
	};

	n.parent = y;
	if (tree.compare(n.data, x.data) < 0) {
		y.left = n;
	} else {
		y.right = n;
	};
	// insert_fixup(tree, node);
	tree.node_count += 1;
};

export fn main() void = {
	fmt::printfln("Yes")!;
	return;
};