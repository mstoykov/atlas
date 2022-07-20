package atlas

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

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

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// from https://stackoverflow.com/a/22892986/5427244
func randSeq() string {
	b := make([]byte, 100)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}
	return string(b)
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
