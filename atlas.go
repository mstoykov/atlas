package atlas

import (
	"sync"
)

type linkKeyType [2]string

type Node struct {
	root    *Node // immutable
	prev    *Node // immutable
	linkKey linkKeyType

	links      map[linkKeyType]*Node // mutable
	linksMutex sync.RWMutex          // TODO replace with atomics if possible
}

func New() *Node {
	n := &Node{
		links: make(map[linkKeyType]*Node),
	}
	n.root = n
	n.prev = n
	return n
}

func (n *Node) ValueByKey(k string) (string, bool) {
	if n.root == n {
		return "", false
	}
	if n.linkKey[0] == k {
		return n.linkKey[1], true
	}
	return n.prev.ValueByKey(k)
}

func (n *Node) Path() map[string]string {
	if n.root == n {
		return make(map[string]string)
	}
	result := n.prev.Path()
	result[n.linkKey[0]] = n.linkKey[1]
	return result
}

func (n *Node) AddLink(key, value string) *Node {
	if n.linkKey[0] == key && n.linkKey[1] == value {
		return n
	}
	n.linksMutex.RLock()
	linkKey := [2]string{key, value}
	result, ok := n.links[linkKey]
	n.linksMutex.RUnlock()
	if ok {
		return result
	}
	result = n.add(linkKey)
	return result
}

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

func (n *Node) add(linkKey linkKeyType) *Node {
	if n.linkKey == linkKey {
		return n
	}
	var newNode *Node
	switch {
	case n.linkKey[0] == linkKey[0]:
		// the key is the same but the value is different
		// we need to add a node at the current "level" to keep the key unique
		// so we make a new node on top of the previous node
		newNode = n.prev.add(linkKey)
	case linkKey[0] < n.linkKey[0] || n.linkKey[0] == "":
		// we just need to add the new node here because
		// we are at the root or we are in the next direct link
		newNode = &Node{
			root:    n.root,
			prev:    n,
			linkKey: linkKey,
			links:   make(map[linkKeyType]*Node),
		}
	default:
		// we need to add this to a previous node in the path
		// then we have to add the current node pair on to the new path
		newNode = n.prev.add(linkKey).add(n.linkKey)
	}
	// the actual adding
	n.linksMutex.Lock()
	if result, ok := n.links[linkKey]; ok {
		// we raced - no problem just return the old
		n.linksMutex.Unlock()
		return result
	}
	n.links[linkKey] = newNode
	n.linksMutex.Unlock()
	return newNode
}
