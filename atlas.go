// Package atlas implement a graph of string to string pairs.
// It is optimzed for when a finite amount of key-value pairs will be constructed in the same way over time.
// It is also thread-safe.
// It's particularly good at taking the same root through the graph over and over again and the Node's can directly be
// used as map keys or compared for equality against each other.
package atlas

import (
	"sync"
)

type linkKeyType [2]string

func (lkt linkKeyType) String() string {
	return "{" + lkt[0] + ": " + lkt[1] + "}"
}

// Node is a node in the atlas
// it is unique per Root and can directly be compared  with `==` to another Node
type Node struct {
	root    *Node       // immutable
	prev    *Node       // immutable
	linkKey linkKeyType // immutable

	links sync.Map
}

// New returns a new root Node. Nodes from different roots being used together have undefined behaviour
func New() *Node {
	n := &Node{}
	n.root = n
	n.prev = n
	return n
}

func (n *Node) String() string {
	if n.IsRoot() {
		return "/"
	}
	return n.prev.String() + "/" + n.linkKey.String()
}

// IsRoot checks if the current Node is the root
func (n *Node) IsRoot() bool {
	return n.root == n
}

// ValueByKey gets the value of key written in this Node or any of it's ancestors
func (n *Node) ValueByKey(k string) (string, bool) {
	if n.root == n {
		return "", false
	}
	if n.linkKey[0] == k {
		return n.linkKey[1], true
	}
	return n.prev.ValueByKey(k)
}

// Path returns a map representing all key/value pairs recorded in the Node
func (n *Node) Path() map[string]string {
	if n.root == n {
		return make(map[string]string)
	}
	result := n.prev.Path()
	result[n.linkKey[0]] = n.linkKey[1]
	return result
}

// AddLink adds the provided key value pair to the tree and returns the new Node that includes all key and values of
// the parent node and the new key value provided. If a key matches it will have the new value provided here.
func (n *Node) AddLink(key, value string) *Node {
	if n.linkKey[0] == key && n.linkKey[1] == value {
		return n
	}
	k, ok := n.links.Load([2]string{key, value})
	if ok {
		return k.(*Node) //nolint:forcetypeassert
	}
	return n.add(key, value)
}

// Contains checks that for each key value in the provided Node there will be the same key with the same value in the
// receiver Node
func (n *Node) Contains(sub *Node) bool {
	if n == sub {
		return true
	}
	if n.root == n {
		return false
	}
	// TODO: https://github.com/mstoykov/atlas/issues/2
	// apparently this is faster than if n.linkKey == sub.linkKey
	if n.linkKey[0] == sub.linkKey[0] && n.linkKey[1] == sub.linkKey[1] {
		return n.prev.Contains(sub.prev)
	}
	return n.prev.Contains(sub)
}

func (n *Node) add(key, value string) *Node {
	if n.linkKey[0] == key && n.linkKey[1] == value {
		return n
	}
	var newNode *Node
	switch {
	case n.linkKey[0] == key:
		// the key is the same but the value is different
		// we need to add a node at the current "level" to keep the key unique
		// so we make a new node on top of the previous node
		newNode = n.prev.add(key, value)
	case key < n.linkKey[0] || n.linkKey[0] == "":
		// we just need to add the new node here because
		// we are at the root or we are in the next direct link
		newNode = &Node{
			root:    n.root,
			prev:    n,
			linkKey: [2]string{key, value},
		}
	default:
		// we need to add this to a previous node in the path
		// then we have to add the current node pair on to the new path
		newNode = n.prev.add(key, value).add(n.linkKey[0], n.linkKey[1])
	}
	// the actual adding
	k, loaded := n.links.LoadOrStore([2]string{key, value}, newNode)
	if loaded {
		// we raced - no problem just return the old
		return k.(*Node) //nolint:forcetypeassert
	}
	return newNode
}
