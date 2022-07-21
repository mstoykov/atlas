package atlas

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodePath(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("c", "6")
	n2 := n1.AddLink("a", "5")
	n3 := n2.AddLink("b", "7")

	require.Equal(t, map[string]string{
		"a": "5",
		"b": "7",
		"c": "6",
	}, n3.Path())
}

func TestNodeAddLink(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("a", "5")
	n2 := n1.AddLink("c", "6")
	n3 := n2.AddLink("b", "7")
	nempty := r.AddLink("", "")

	require.Equal(t, n3.Path(), map[string]string{
		"a": "5",
		"b": "7",
		"c": "6",
	})
	require.True(t, r != n1)
	require.True(t, r != n2)
	require.True(t, r != n3)
	require.True(t, n2 != n1)
	require.True(t, n3 != n1)
	require.True(t, n2 != n3)
	require.True(t, r == nempty)
	require.True(t, n2 == r.AddLink("c", "6").AddLink("a", "5"))
	require.True(t, n2 == r.AddLink("c", "6").AddLink("a", "5").AddLink("a", "5"))
}

func TestNodeValueByKey(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("c", "6").AddLink("a", "5")
	n2 := r.AddLink("b", "7")

	v, ok := n1.ValueByKey("c")
	require.True(t, ok)
	assert.Equal(t, "6", v)

	_, ok = n1.ValueByKey("b")
	require.False(t, ok)

	v, ok = n2.ValueByKey("b")
	require.True(t, ok)
	assert.Equal(t, "7", v)
}

func TestNodeContains(t *testing.T) {
	t.Parallel()

	r := New()

	n1 := r.AddLink("a", "5").
		AddLink("c", "6").
		AddLink("b", "7")

	n2 := r.AddLink("b", "7").AddLink("c", "6")
	n3 := r.AddLink("b", "7").AddLink("c", "4")
	n4 := r.AddLink("a", "5")

	n5 := r.AddLink("d", "9").
		AddLink("b", "7").
		AddLink("c", "4")

	assert.True(t, r.Contains(r))   // {} | {}
	assert.True(t, n1.Contains(n1)) // A5,C6,B7 | A5,C6,B7

	assert.True(t, n2.Contains(r))    // B7,C6 | {}
	require.False(t, n2.Contains(n1)) // B7,C6 | A5,C6,B7
	assert.False(t, r.Contains(n2))   // {} | B7,C6

	require.True(t, n1.Contains(n4))  // A5,C6,B7 | A5
	require.True(t, n1.Contains(n2))  // A5,C6,B7 | B7,C6
	require.False(t, n1.Contains(n3)) // A5,C6,B7 | B7,C4

	require.False(t, n3.Contains(n5)) // B7,C4 | A5,C6,B7
	require.False(t, n3.Contains(n5)) // B7,C4 | D9,B7,C4
	require.False(t, n3.Contains(n2)) // B7,C4 | B7,C6
}

func TestNodeIsRoot(t *testing.T) {
	t.Parallel()

	r := New()
	assert.True(t, r.IsRoot())
	subnode := r.AddLink("key1", "val1")
	assert.False(t, subnode.IsRoot())
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// from https://stackoverflow.com/a/22892986/5427244
func randSeq() string {
	b := make([]byte, 100)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}
	return string(b)
}

// test helper for pseudo-reproducibly-random and concurrent atlas tests
type testHelper struct {
	t    *testing.T
	rand *rand.Rand

	tags           []linkKeyType
	tagKeys        []string
	tagValuesByKey map[string][]string
}

func newTestHelper(t *testing.T, keyCountFrom, keyCountTo, valuesPerKeyFrom, valuesPerKeyTo int) *testHelper {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed)) //nolint:gosec

	th := &testHelper{
		t:    t,
		rand: r,
	}

	tagKeyCount := th.randInt(keyCountFrom, keyCountTo)
	t.Logf("RandomSource=%d, tagKeyCount=%d\n", seed, tagKeyCount)

	th.tagKeys = make([]string, tagKeyCount)
	th.tagValuesByKey = make(map[string][]string, tagKeyCount)
	for i := 0; i < tagKeyCount; i++ {
		key := th.randString(5, 30)
		th.tagKeys[i] = key

		valuesForKeyCount := th.randInt(valuesPerKeyFrom, valuesPerKeyTo)
		values := make([]string, valuesForKeyCount)
		t.Logf("Generating %d tag values for tag '%s'...", valuesForKeyCount, key)
		for j := 0; j < valuesForKeyCount; j++ {
			value := th.randString(10, 100)
			values[j] = value
			th.tags = append(th.tags, linkKeyType{key, value})
		}
		th.tagValuesByKey[key] = values
	}

	t.Logf("Generated %d tag pairs in total: %q", len(th.tags), th.tagValuesByKey)
	return th
}

func (th *testHelper) randInt(from, to int) int {
	return from + th.rand.Intn(to-from)
}

func (th *testHelper) randString(minLen, maxLen int) string {
	b := make([]byte, th.randInt(minLen, maxLen))
	for i := range b {
		b[i] = letters[th.randInt(0, len(letters))]
	}
	return string(b)
}

// https://en.wikipedia.org/wiki/Euclidean_algorithm#Implementations
func (th *testHelper) areCoPrime(a, b int) bool {
	var t int
	for b != 0 {
		t = b
		b = a % b
		a = t
	}
	return a == 1
}

// https://lemire.me/blog/2017/09/18/visiting-all-values-in-an-array-exactly-once-in-random-order/
func (th *testHelper) getRandCoPrimeAndOffset(num int) (int, int) {
	randOffset := th.randInt(2, 10*num) // doesn't really matter
	th.t.Logf("Generating a random co-prime and offset (%d) to iterate pseudo-randomly over %d elements", randOffset, num)
	for {
		c := th.randInt(1+num/2, 10*num) // doesn't really matter
		th.t.Logf("Checking to see if %d and %d are co-prime...", num, c)
		if th.areCoPrime(num, c) {
			return c, randOffset
		}
	}
}

func (th *testHelper) getOneTagPerKey(maxKeys int) []linkKeyType {
	keysLen := len(th.tagKeys)
	res := make([]linkKeyType, maxKeys)

	step, offset := th.getRandCoPrimeAndOffset(keysLen)
	for i := 0; i < maxKeys; i++ {
		// https://lemire.me/blog/2017/09/18/visiting-all-values-in-an-array-exactly-once-in-random-order/
		key := th.tagKeys[(i*step+offset)%keysLen]
		values := th.tagValuesByKey[key]
		res[i] = linkKeyType{key, values[th.randInt(0, len(values))]}
	}

	return res
}

func testFinalNodesEquality(th *testHelper, startingNodes []*Node, tags []linkKeyType) {
	tagsLen := len(tags)
	vusCount := th.randInt(20, 50)
	vusReady := &sync.WaitGroup{}
	vusReady.Add(vusCount)
	startVUs := make(chan struct{})
	vusDone := &sync.WaitGroup{}
	vusDone.Add(vusCount)

	finalNodes := make([]*Node, vusCount)
	for i := 0; i < vusCount; i++ {
		randStep, randOffset := th.getRandCoPrimeAndOffset(tagsLen)
		randStartingNode := th.randInt(0, len(startingNodes))
		th.t.Logf(
			"VU %d will use starting Node %d, co-prime %d, offset %d to iterate over %d elements",
			i, randStartingNode, randStep, randOffset, tagsLen,
		)
		go func(vuID, startNode, step, offset int) {
			vusReady.Done()
			<-startVUs // start all VUs as simultaneously as possible
			n := startingNodes[startNode]
			myNodes := make([]*Node, tagsLen)
			for j := 0; j < tagsLen; j++ {
				tagIndex := (j*randStep + offset) % tagsLen
				// th.t.Logf("VU %03d iterates over index %03d (%s)", vuID, tagIndex, tags[tagIndex])
				n = n.add(tags[tagIndex][0], tags[tagIndex][1])
				myNodes[j] = n
			}
			finalNodes[vuID] = n
			vusDone.Done()
		}(i, randStartingNode, randStep, randOffset)
	}

	vusReady.Wait()
	close(startVUs)
	vusDone.Wait()

	for i := 1; i < vusCount; i++ {
		if finalNodes[0] != finalNodes[i] {
			th.t.Errorf("Final node 0 is not equal to final node %d:\n\t%s\n\t%s", i, finalNodes[0], finalNodes[i])
		}
	}
}

func getAllNodes(root *Node) []*Node {
	seen := map[*Node]struct{}{}
	nodes := []*Node{}
	var addNodes func(n *Node)
	addNodes = func(n *Node) {
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		nodes = append(nodes, n)
		n.links.Range(func(key, value any) bool {
			addNodes(value.(*Node)) //nolint:forcetypeassert
			return true
		})
	}
	addNodes(root)
	return nodes
}

func TestFinalNodesEquality(t *testing.T) {
	t.Parallel()

	th := newTestHelper(t, 10, 30, 5, 50)

	root := New()
	// Starting from the root, if we add the same tag keys and values, but in
	// different order, we expect to arrive at the same final node
	subsetTestCount := th.randInt(15, 30)
	tagSets := make([][]linkKeyType, subsetTestCount)
	// We generate the tags sets before we test with VUs, so the RNG numbers
	// that influence the key and tag picking are not affected by VU counts
	for i := 0; i < subsetTestCount; i++ {
		keysCount := th.randInt(len(th.tagKeys)/2, len(th.tagKeys)) // do not generally use all keys
		tagSets[i] = th.getOneTagPerKey(keysCount)
	}
	for i := 0; i < subsetTestCount; i++ {
		t.Logf(
			"Final node equality test with %d/%d with tags for %d keys: %q",
			i+1, subsetTestCount, len(tagSets[i]), tagSets[i],
		)
		testFinalNodesEquality(th, []*Node{root}, tagSets[i])
	}

	// Add all of the observed nodes in the set of starting nodes, because...
	t.Logf("Gathering all graph nodes...")
	startTime := time.Now()
	startingNodes := getAllNodes(root)
	t.Logf("Gathered %d starting nodes for %s!", len(startingNodes), time.Since(startTime))

	// ... if we start at any node, but we make sure to add exactly one tag with
	// every key, we still expect to arrive at the same final node, even if we
	// add the tags in a different order from every VU
	allTagsTestCount := th.randInt(15, 30)
	for i := 0; i < allTagsTestCount; i++ {
		tags := th.getOneTagPerKey(len(th.tagKeys)) // use tags from all keys
		t.Logf("Final node equality test with %d/%d with tags for all keys: %q", i+1, subsetTestCount, tags)
		testFinalNodesEquality(th, startingNodes, tags)
	}
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	// this just test that adding stuff is not racy
	r := New()
	values := make([]string, 10000)
	keys := make([]string, 15)
	for i := 0; i < len(keys); i++ {
		keys[i] = randSeq()
	}
	for i := 0; i < len(values); i++ {
		values[i] = randSeq()
	}
	concurrency := 128
	repetitions := 10240
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func(wid int) {
			defer wg.Done()
			n := r
			for j := 0; j < repetitions; j++ {
				index := wid + j
				n = n.AddLink(keys[index%len(keys)], values[index%len(keys)])
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkNodeConcurrencyBad(b *testing.B) {
	ixrand := func(nvals int) int {
		return rand.Int() % nvals //nolint:gosec
	}
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			keys := make([]string, 15)
			for i := 0; i < len(keys); i++ {
				keys[i] = randSeq()
			}
			for i := 0; i < len(values); i++ {
				values[i] = randSeq()
			}
			r := New()
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					r.AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					)
				}
			})
		})
	}
}

func BenchmarkNodeRealistic(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			for i := 0; i < len(values); i++ {
				values[i] = randSeq()
			}
			r := New().
				AddLink("labelone", "valueone").
				AddLink("labeltthree", "valuetthree").
				AddLink("labelfour", "valuefour").
				AddLink("labelfive", "valuefive").
				AddLink("labelsix", "valuefive").
				AddLink("labelseven", "valuefive").
				AddLink("labeleigth", "valuefive").
				AddLink("labeltwo", "valuetwo")
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				i := 0
				for p.Next() {
					i++
					n := r.AddLink(
						"badkey",
						values[i%len(values)],
					)
					if i%2 == 0 {
						n = n.AddLink("labelsix", "someOtheStrangeValue")
					}
					switch i % 7 {
					case 0, 1, 2:
						n.AddLink("okayLabel", "200")
					case 3, 4:
						n.AddLink("okayLabel", "400")
					case 5:
						n.AddLink("okayLabel", "500")
					case 6:
						n.AddLink("okayLabel", "0")
					}
				}
			})
		})
	}
}

func BenchmarkNodeContains(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			for i := 0; i < len(values); i++ {
				values[i] = randSeq()
			}
			r := New()
			n := r.AddLink("labelone", "valueone").
				AddLink("labeltwo", "valuetwo").
				AddLink("labeltthree", "valuetthree").
				AddLink("labelfour", "valuefour").
				AddLink("labelfive", "valuefive").
				AddLink("labelsix", "valuefive").
				AddLink("labelseven", "valuefive").
				AddLink("labeleigth", "valuefive")

			n2 := r.AddLink("labelone", "valueone").
				AddLink("labeltthree", "valuetthree").
				AddLink("labelfour", "valuefour").
				AddLink("labelfive", "valuefive").
				AddLink("labelsix", "valuefive").
				AddLink("labelseven", "valuefive").
				AddLink("labeleigth", "valuefive")

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					n.Contains(n2)
				}
			})
		})
	}
}
