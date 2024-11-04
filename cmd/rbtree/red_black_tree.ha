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

fn insert_fixup(tree: *tree, n: *node) void = {
	let root: *node = match(tree.root) {
	case null =>
		return;
	case let n: *node =>
		yield n;
	};

	for (n.parent.color == color::RED) {
		if (n.parent == n.parent.parent.left) {
			let parent_sibling = n.parent.parent.right;
			if (parent_sibling.color == color::RED)
			{
				n.parent.color = color::BLACK;
				parent_sibling.color = color::BLACK;
				n.parent.parent.color = color::RED;
				n = n.parent.parent;
			} else {
				if (n == n.parent.right) {
					n = n.parent;
					left_rotate(tree, n);
				};
				n.parent.color = color::BLACK;
				n.parent.parent.color = color::RED;
				right_rotate(tree, n.parent.parent);
			};
		} else {
			let parent_sibling = n.parent.parent.left;
			if (parent_sibling.color == color::RED) {
				n.parent.color = color::BLACK;
				parent_sibling.color = color::BLACK;
				n.parent.parent.color = color::RED;
				n = n.parent.parent;
			} else {
				if (n == n.parent.left) {
					n = n.parent;
					right_rotate(tree, n);
				};
				n.parent.color = color::BLACK;
				n.parent.parent.color = color::RED;
				left_rotate(tree, n.parent.parent);
			};
		};
	};
	root.color = color::BLACK;
};

fn left_rotate(tree: *tree, n: *node) void = {
	let sibling = n.right;

	n.right = sibling.left;
	if (sibling.left != &tree.nil) {
		sibling.left.parent = n;
	};
	sibling.parent = n.parent;
	if (sibling.parent == &tree.nil) {
		tree.root = sibling;
	} else if (n == sibling.parent.left) {
		n.parent.left = sibling;
	} else {
		n.parent.right = sibling;
	};
	sibling.left = n;
	n.parent = sibling;
};

fn right_rotate(tree: *tree, n: *node) void = {
	let sibling = n.left;

	n.left = sibling.right;
	if (sibling.right != &tree.nil) {
		sibling.right.parent = n;
	};
	sibling.parent = n.parent;
	if (sibling.parent == &tree.nil) {
		tree.root = sibling;
	} else if (n == sibling.parent.left) {
		n.parent.left = sibling;
	} else {
		n.parent.right = sibling;
	};
	sibling.right = n;
	n.parent = sibling;
};

export fn minimum(tree: *tree) *node = {
	let current: *node = match(tree.root) {
	case null =>
		return &tree.nil;
	case let n: *node =>
		yield n;
	};

	return minimum_subtree(tree, current);
};

export fn maximum(tree: *tree) *node = {
	let current: *node = match(tree.root) {
	case null =>
		return &tree.nil;
	case let n: *node =>
		yield n;
	};

	return maximum_subtree(tree, current);
};

export fn minimum_subtree(tree: *tree, n: *node) *node = {
	let current = n;

	if (current == &tree.nil) {
		return current;
	};
	for (current.left != &tree.nil) {
		current = current.left;
	};
	return current;
};

export fn maximum_subtree(tree: *tree, n: *node) *node = {
	let current = n;

	if (current == &tree.nil) {
		return current;
	};
	for (current.right != &tree.nil) {
		current = current.right;
	};
	return current;
};

export fn get(tree: *tree, data: data) data = {

};

export fn del(tree: *tree, to_del: data) bool = {

};

fn delete_node(tree: *tree, to_be_deleted: *node) void = {
	let successor = to_be_deleted;
	let to_be_fixed = &tree.nil;
	let successor_original_color = successor.color;

	if (to_be_deleted.left == &tree.nil) {
		to_be_fixed = to_be_deleted.right;
		transplant(tree, to_be_deleted, to_be_fixed);
	} else if (to_be_deleted.right == &tree.nil) {
		to_be_fixed = to_be_deleted.left;
		transplant(tree, to_be_deleted, to_be_fixed);
	} else {
		successor = minimum_subtree(tree, to_be_deleted.right);
		successor_original_color = successor.color;
		to_be_fixed = successor.right;
		if (successor.parent == to_be_deleted) {
			to_be_fixed.parent = successor;
		} else {
			transplant(tree, successor, successor.right);
			successor.right = to_be_deleted.right;
			successor.right.parent = successor;
		};
		transplant(tree, to_be_deleted, successor);
		successor.left = to_be_deleted.left;
		successor.left.parent = successor;
		successor.color = to_be_deleted.color;
	};

	if (successor_original_color == color::BLACK) {
		delete_fixup(tree, to_be_fixed);
	};
};

fn delete_fixup(tree: *tree, n: *node) void = {
	let root: *node = match(tree.root) {
	case null =>
		fmt::fatalf("Issue with delete_fixup idk");
		// return;
	case let n: *node =>
		yield n;
	};

	for (n != root && n.color == color::BLACK) {
		if (n == n.parent.left) {
			let sibling = n.parent.right;
			if (sibling.color == color::RED) {
				sibling.color = color::BLACK;
				n.parent.color = color::RED;
				left_rotate(tree, n.parent);
				sibling = n.parent.right;
			};

			if (sibling.left.color == color::BLACK
			&& sibling.right.color == color::BLACK) {
				sibling.color = color::RED;
				n = n.parent;
			} else if (sibling.right.color == color::BLACK) {
				sibling.left.color = color::BLACK;
				sibling.color = color::RED;
				right_rotate(tree, sibling);
				sibling = n.parent.right;
			} else {
				sibling.color = n.parent.color;
				n.parent.color = color::BLACK;
				sibling.right.color = color::BLACK;
				left_rotate(tree, n.parent);
				n = root;
			};
		} else {
			let sibling = n.parent.left;
			if (sibling.color == color::RED) {
				sibling.color = color::BLACK;
				n.parent.color = color::RED;
				right_rotate(tree, n.parent);
				sibling = n.parent.left;
			};

			if (sibling.left.color == color::BLACK
			&& sibling.right.color == color::BLACK) {
				sibling.color = color::RED;
				n = n.parent;
			} else if (sibling.left.color == color::BLACK) {
				sibling.right.color = color::BLACK;
				sibling.color = color::RED;
				left_rotate(tree, sibling);
				sibling = n.parent.left;
			} else {
				sibling.color = n.parent.color;
				n.parent.color = color::BLACK;
				sibling.left.color = color::BLACK;
				right_rotate(tree, n.parent);
				n = root;
			};
		};
	};
};

fn transplant(tree: *tree, u: *node, v: *node) void = {
	if (u.parent == &tree.nil) {
		tree.root = v;
	} else if (u == u.parent.left) {
		u.parent.left = v;
	} else {
		u.parent.right = v;
	};
	v.parent = u.parent;
};

export fn main() void = {
	fmt::printfln("Yes")!;
	return;
};