# Intro

A PoC connected graph of string to string pairs (map[string]string) with amortized construction costs and O(1) equality
cost.

The O(1) equality comes from comparing the pointers to the struct instead of the struct itself. 

Amortized cost mean that as time pass on and you stop adding more pairs the fairly high initial costs go down.

Additionally this should considerably reduce GC [citation needed] as we don't continuously make new objects.


# TODO 
- Add an operation to check if a Node has given key-value pair(s), efficiently - preferably at O(k) where k is the length of the path for the Node
- Try to make the map less of congestion point. Maybe before trying to make our own, sync.Map or a map per key or maps per key, a trie for the values which will be the more high cordiality ones :shrug:
- count the number of nodes, for example in the root :shrug:
- Better names `GoTo` was called `Set` but both seem terrible. 
- (in the future) do not have any of this in the golang memory as this just adds to things the GC needs to check but will in practice never free as we keep this memory around forever.
- (in further in the future) try to have a GC for nodes that haven't been accessed in a long while :shrug:
