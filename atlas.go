package atlas

import (
	"sync"
)

type Node struct {
	root    *Node // immutable
	prev    *Node // immutable
	linkKey linkKeyType

	links      map[linkKeyType]*Node // mutable
	linksMutex sync.RWMutex          // TODO replace with atomics if possible
}

type linkKeyType [2]string

func (n *Node) GoTo(key, value string) *Node {
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

func (n *Node) add(linkKey linkKeyType) *Node {
	if n.linkKey == linkKey {
		return n
	}
	var newNode *Node
	switch {
	case n.linkKey[0] == linkKey[0]:
		// the key is save the value is different
		// make a new node from the prev
		newNode = n.prev.add(linkKey)
	case linkKey[0] < n.linkKey[0] || n.linkKey[0] == "":
		// we just need to add it here or we at the root
		newNode = &Node{
			root:    n.root,
			prev:    n,
			linkKey: linkKey,
			links:   make(map[linkKeyType]*Node),
		}
	default:
		// we need to add this to a previous node in the path to the root
		// and then add the current node keys on top of that
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

func (n *Node) Path() map[string]string {
	if n.root == n {
		return make(map[string]string)
	}
	result := n.prev.Path()
	result[n.linkKey[0]] = n.linkKey[1]
	return result
}

func New() *Node {
	n := &Node{
		links: make(map[linkKeyType]*Node),
	}
	n.root = n
	n.prev = n
	return n
}
