package atlas

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	r := New()
	n1 := r.GoTo("a", "5")
	n2 := n1.GoTo("c", "6")
	n3 := n2.GoTo("b", "7")
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
	require.True(t, n2 == r.GoTo("c", "6").GoTo("a", "5"))
	require.True(t, n2 == r.GoTo("c", "6").GoTo("a", "5").GoTo("a", "5"))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// from https://stackoverflow.com/a/22892986/5427244
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestConcurrency(t *testing.T) {
	// this just test that adding stuff is not racy
	r := New()
	values := make([]string, 10000)
	keys := make([]string, 15)
	for i := 0; i < len(keys); i++ {
		keys[i] = randSeq(100)
	}
	for i := 0; i < len(values); i++ {
		values[i] = randSeq(100)
	}
	concurrency := 128
	repetitions := 10240
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func(j int) {
			defer wg.Done()
			n := r
			for i := 0; i < repetitions; i++ {
				index := j + i
				n = n.GoTo(keys[index%len(keys)], values[index%len(keys)])
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkConcurrencyBad(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			keys := make([]string, 15)
			for i := 0; i < len(keys); i++ {
				keys[i] = randSeq(100)
			}
			for i := 0; i < len(values); i++ {
				values[i] = randSeq(100)
			}
			r := New()
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					r.GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					).GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					).GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					).GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					).GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					).GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					).GoTo(
						keys[rand.Int()%len(keys)],
						values[rand.Int()%len(values)],
					)
				}
			})
		})
	}
}

func BenchmarkRealistic(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			for i := 0; i < len(values); i++ {
				values[i] = randSeq(100)
			}
			r := New().
				GoTo("labelone", "valueone").
				GoTo("labeltthree", "valuetthree").
				GoTo("labelfour", "valuefour").
				GoTo("labelfive", "valuefive").
				GoTo("labelsix", "valuefive").
				GoTo("labelseven", "valuefive").
				GoTo("labeleigth", "valuefive").
				GoTo("labeltwo", "valuetwo")
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				i := 0
				for p.Next() {
					i++
					n := r.GoTo(
						"badkey",
						values[i%len(values)],
					)
					if i%2 == 0 {
						n = n.GoTo("labelsix", "someOtheStrangeValue")
					}
					switch i % 7 {
					case 0, 1, 2:
						n.GoTo("okayLabel", "200")
					case 3, 4:
						n.GoTo("okayLabel", "400")
					case 5:
						n.GoTo("okayLabel", "500")
					case 6:
						n.GoTo("okayLabel", "0")
					}

				}
			})
		})
	}
}
