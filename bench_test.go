package atlas

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkStaticAndNoConcurrency(b *testing.B) {
	b.Run("SliceFromMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tags := FromMap(map[string]string{
				"labelone":    "valueone",
				"labeltthree": "valuetthree",
				"labelfour":   "valuefour",
				"labelfive":   "valuefive",
				"labelsix":    "valuefive",
				"labelseven":  "valuefive",
				"labeleigth":  "valuefive",
				"labeltwo":    "valuetwo",
			})
			h := tags.Hash()
			_ = h
		}
	})

	b.Run("SliceAdding", func(b *testing.B) {
		tagSetPool := sync.Pool{
			New: func() any {
				return NewTagSet()
			},
		}
		for i := 0; i < b.N; i++ {
			tags := tagSetPool.Get().(*TagSet)
			tags.Reset()
			tags.AddTag("labelone", "valueone")
			tags.AddTag("labeltthree", "valuetthree")
			tags.AddTag("labelfour", "valuefour")
			tags.AddTag("labelfive", "valuefive")
			tags.AddTag("labelsix", "valuefive")
			tags.AddTag("labelseven", "valuefive")
			tags.AddTag("labeleigth", "valuefive")
			tags.AddTag("labeltwo", "valuetwo")
			tags.Sort()
			h := tags.Hash()
			_ = h

			tagSetPool.Put(tags)
		}
	})

	b.Run("Atlas", func(b *testing.B) {
		r := New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.AddLink("labelone", "valueone").
				AddLink("labeltthree", "valuetthree").
				AddLink("labelfour", "valuefour").
				AddLink("labelfive", "valuefive").
				AddLink("labelsix", "valuefive").
				AddLink("labelseven", "valuefive").
				AddLink("labeleigth", "valuefive").
				AddLink("labeltwo", "valuetwo")
		}
	})
}

func BenchmarkStaticConcurrentAtlas(b *testing.B) {
	for _, n := range []int{runtime.GOMAXPROCS(0), 1000, 10000, 50000} {
		b.Run(strconv.Itoa(n)+"gos", func(b *testing.B) {
			r := New()

			b.SetParallelism(n)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					r.AddLink("labelone", "valueone").
						AddLink("labeltthree", "valuetthree").
						AddLink("labelfour", "valuefour").
						AddLink("labelfive", "valuefive").
						AddLink("labelsix", "valuefive").
						AddLink("labelseven", "valuefive").
						AddLink("labeleigth", "valuefive").
						AddLink("labeltwo", "valuetwo")
				}
			})
		})
	}
}

func BenchmarkStaticConcurrentSlice(b *testing.B) {
	for _, n := range []int{runtime.GOMAXPROCS(0), 1000, 10000, 50000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.SetParallelism(n)
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					//tags := FromMap(map[string]string{
					//"labelone":    "valueone",
					//"labeltthree": "valuetthree",
					//"labelfour":   "valuefour",
					//"labelfive":   "valuefive",
					//"labelsix":    "valuefive",
					//"labelseven":  "valuefive",
					//"labeleigth":  "valuefive",
					//"labeltwo":    "valuetwo",
					//})
					//h := tags.Hash()
					//_ = h

					tags := NewTagSet()
					tags.AddTag("labelone", "valueone")
					tags.AddTag("labeltthree", "valuetthree")
					tags.AddTag("labelfour", "valuefour")
					tags.AddTag("labelfive", "valuefive")
					tags.AddTag("labelsix", "valuefive")
					tags.AddTag("labelseven", "valuefive")
					tags.AddTag("labeleigth", "valuefive")
					tags.AddTag("labeltwo", "valuetwo")
					tags.Sort()
					h := tags.Hash()
					_ = h
				}
			})
		})
	}
}

type dynamicBenchData struct {
	values []string
	keys   []string

	// It stores a series of random generated indexes
	// to use to access in random order to a slice without
	// compute them during the benchmark
	randomIndex []uint64

	maxItemIndex  uint64
	maxItemKeys   uint64
	maxItemValues uint64
}

func newDynamicBenchData() dynamicBenchData {
	bdata := dynamicBenchData{
		randomIndex: make([]uint64, 1000),
		values:      make([]string, 1000),
		keys:        make([]string, 15),
	}
	for i := 0; i < len(bdata.keys); i++ {
		bdata.keys[i] = randSeq()
	}

	for i := 0; i < len(bdata.values); i++ {
		bdata.values[i] = randSeq()
		bdata.randomIndex[i] = uint64(rand.Int())
	}

	bdata.maxItemIndex = uint64(len(bdata.randomIndex) - 1)
	bdata.maxItemKeys = uint64(len(bdata.keys) - 1)
	bdata.maxItemValues = uint64(len(bdata.values) - 1)

	return bdata
}

// It returns an index to use for accessing the keys slice in a random order
func (bdata *dynamicBenchData) RandKey(counter uint64) uint64 {
	return bdata.randomIndex[counter%bdata.maxItemIndex] % bdata.maxItemKeys
}

// It returns an index to use for accessing the values slice in a random order
func (bdata *dynamicBenchData) RandValue(counter uint64) uint64 {
	return bdata.randomIndex[counter%bdata.maxItemIndex] % bdata.maxItemValues
}

func BenchmarkDynamic(b *testing.B) {
	// variance(0%, 25%, 50%, 75%, 100%)
	for _, variance := range []uint64{0, 2, 4, 6, 8} {
		for _, n := range []int{runtime.GOMAXPROCS(0), 100, 1000, 10000, 50000} {
			name := fmt.Sprintf("Var%d%%", variance*100/8)
			if n >= 1000 {
				name += fmt.Sprintf("G%dk", n/1000)
			} else {
				name += fmt.Sprintf("G%d", n)
			}
			b.Run(name, func(b *testing.B) {
				bdata := newDynamicBenchData()

				tags := []Tag{
					{"labelone", "valueone"},
					{"labeltwo", "valuetwo"},
					{"labelthree", "valuetthree"},
					{"labelfour", "valuefour"},
					{"labelfive", "valuefive"},
					{"labelsix", "valuefive"},
					{"labelseven", "valuefive"},
					{"labeleigth", "valuefive"},
				}

				b.Run("Node", func(b *testing.B) {
					r := New()
					b.ReportAllocs()
					b.ResetTimer()
					b.SetParallelism(n)
					b.RunParallel(func(p *testing.PB) {
						cycles := uint64(0)
						for p.Next() {
							for i := uint64(0); i < 8; i++ {
								if i < variance {
									r.AddLink(
										bdata.keys[bdata.RandKey(cycles+i)],
										bdata.values[bdata.RandValue(cycles+i)],
									)
								} else {
									r.AddLink(tags[i].Key, tags[i].Value)
								}
							}
							cycles++
						}
					})
				})

				b.Run("Slice", func(b *testing.B) {
					tagSetPool := sync.Pool{
						New: func() any {
							return NewTagSet()
						},
					}

					b.ReportAllocs()
					b.ResetTimer()
					b.SetParallelism(n)

					b.RunParallel(func(p *testing.PB) {
						cycles := uint64(0)
						for p.Next() {
							r := tagSetPool.Get().(*TagSet)
							r.Reset()

							for i := uint64(0); i < 8; i++ {
								if i < variance {
									r.AddTag(
										bdata.keys[bdata.RandKey(cycles+i)],
										bdata.values[bdata.RandValue(cycles+i)],
									)
								} else {
									r.AddTag(tags[i].Key, tags[i].Value)
								}
							}
							r.Sort()
							h := r.Hash()
							_ = h

							tagSetPool.Put(r)
							cycles++
						}
					})
				})
			})
		}
	}
}

// There is an open conflict - O(N) where N = len(n2)
// - When they are the equal Atlas can just exec a pointer comparison so faster, but this is not so relevant because probably we already executed IsEqual before and we don't need this comparison.
// - When they are different:
//		- if n1 is bigger then Atlas is faster based on its faster path resolution.
//		- if n2 is bigger than n1 then the Slice is faster because it returns after it has checked the len()s.
func BenchmarkContains(b *testing.B) {
	b.Run("Atlas", func(b *testing.B) {
		r := New()

		n := r.AddLink("labelone", "valueone").
			AddLink("labelthree", "valuethree").
			AddLink("labelfour", "valuefour").
			AddLink("labelfive", "valuefive").
			AddLink("labelsix", "valuefive").
			AddLink("labelseven", "valuefive").
			AddLink("labeleigth", "valuefive")
			// AddLink("labeltwo", "valuetwo")

		n2 := r.AddLink("labelone", "valueone").
			AddLink("labeltwo", "valuetwo").
			AddLink("labelthree", "valuethree").
			AddLink("labelfour", "valuefour").
			AddLink("labelfive", "valuefive").
			AddLink("labelsix", "valuefive").
			AddLink("labelseven", "valuefive").
			AddLink("labeleigth", "valuefive")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			assert.False(b, n.Contains(n2))
		}
	})

	b.Run("Slice", func(b *testing.B) {
		t1 := FromMap(map[string]string{
			"labelone":   "valueone",
			"labelthree": "valuethree",
			"labelfour":  "valuefour",
			"labelfive":  "valuefive",
			"labelsix":   "valuefive",
			"labelseven": "valuefive",
			"labeleigth": "valuefive",
			//"labeltwo":   "valuetwo",
		})
		t2 := FromMap(map[string]string{
			"labelone":   "valueone",
			"labeltwo":   "valuetwo",
			"labelthree": "valuethree",
			"labelfour":  "valuefour",
			"labelfive":  "valuefive",
			"labelsix":   "valuefive",
			"labelseven": "valuefive",
			"labeleigth": "valuefive",
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			assert.False(b, t1.Contains(t2))
		}
	})
}

// Mostly the same
func BenchmarkDirectComparison(b *testing.B) {
	b.Run("Atlas", func(b *testing.B) {
		t1 := FromMap(map[string]string{
			"labelfive":  "valuefive",
			"labelone":   "valueone",
			"labelthree": "valuethree",
			"labelfour":  "valuefour",
			"labelsix":   "valuefive",
			"labelseven": "valuefive",
			"labeleigth": "valuefive",
			"labeltwo":   "valuetwo",
		})
		t2 := FromMap(map[string]string{
			"labelone":   "valueone",
			"labeltwo":   "valuetwo",
			"labelthree": "valuethree",
			"labelfour":  "valuefour",
			"labelfive":  "valuefive",
			"labelsix":   "valuefive",
			"labelseven": "valuefive",
			"labeleigth": "valuefive",
		})
		require.Equal(b, t1.Keys(), t2.Keys())
		h1 := t1.Hash()
		h2 := t2.Hash()
		require.Equal(b, h1, h2)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = h1 == h2
		}
	})
	b.Run("Slice", func(b *testing.B) {
		r := New()

		n1 := r.AddLink("labelone", "valueone").
			AddLink("labelthree", "valuethree").
			AddLink("labelfour", "valuefour").
			AddLink("labelfive", "valuefive").
			AddLink("labelsix", "valuefive").
			AddLink("labelseven", "valuefive").
			AddLink("labeleigth", "valuefive").
			AddLink("labeltwo", "valuetwo")

		n2 := r.AddLink("labelone", "valueone").
			AddLink("labeltwo", "valuetwo").
			AddLink("labelthree", "valuethree").
			AddLink("labelfour", "valuefour").
			AddLink("labelfive", "valuefive").
			AddLink("labelsix", "valuefive").
			AddLink("labelseven", "valuefive").
			AddLink("labeleigth", "valuefive")

		require.True(b, n1 == n2)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = n1 == n2
		}
	})
}
