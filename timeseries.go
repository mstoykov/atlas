package atlas

import (
	"sort"

	"github.com/cespare/xxhash/v2"
)

type Tag struct {
	Key, Value string
}

type TagSet struct {
	tags     []Tag
	tagIndex map[string]int
}

func (ts *TagSet) Reset() {
	ts.tags = ts.tags[:0]
	ts.tagIndex = make(map[string]int)
}

func (ts *TagSet) Len() int           { return len(ts.tags) }
func (ts *TagSet) Less(i, j int) bool { return ts.tags[i].Key < ts.tags[j].Key }
func (ts *TagSet) Swap(i, j int)      { ts.tags[i], ts.tags[j] = ts.tags[j], ts.tags[i] }

func FromMap(m map[string]string) *TagSet {
	ts := &TagSet{
		tags:     make([]Tag, 0, len(m)),
		tagIndex: make(map[string]int, len(m)),
	}
	for k, v := range m {
		ts.tags = append(ts.tags, Tag{Key: k, Value: v})
	}
	ts.Sort()
	return ts
}

func NewTagSet() *TagSet {
	return &TagSet{
		tags:     nil,
		tagIndex: make(map[string]int),
	}
}

func (ts *TagSet) AddTag(key, value string) {
	ts.tags = append(ts.tags, Tag{Key: key, Value: value})
}

func (ts *TagSet) Sort() {
	sort.Sort(ts)
	for i := 0; i < len(ts.tags); i++ {
		ts.tagIndex[ts.tags[i].Key] = i
	}
}

func (ts *TagSet) Hash() uint64 {
	hasher := xxhash.New()
	for i := 0; i < len(ts.tags); i++ {
		hasher.WriteString(ts.tags[i].Key)
		hasher.Write(bytesep)
		hasher.WriteString(ts.tags[i].Value)
		hasher.Write(bytesep)
	}

	h := hasher.Sum64()
	hasher.Reset()
	return h
}

func (ts *TagSet) Keys() []string {
	ks := make([]string, len(ts.tags))
	for i := 0; i < len(ts.tags); i++ {
		ks[i] = ts.tags[i].Key
	}
	return ks
}

func (ts *TagSet) Contains(other *TagSet) bool {
	if ts == other || other == nil {
		return true
	}
	if len(ts.tags) < len(other.tags) {
		return false
	}

	for _, ot := range other.tags {
		if ix, ok := ts.tagIndex[ot.Key]; !ok || ts.tags[ix].Value != ot.Value {
			return false
		}
	}

	return true
}

var bytesep = []byte{0xff}
