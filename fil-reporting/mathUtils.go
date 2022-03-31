package fil_reporting

import (
	"fil-benchmark/datastructures"
	"math"
	"sort"
)

func Map(vs []datastructures.BlockLog, f func(datastructures.BlockLog) uint64) []uint64 {
	vsm := make([]uint64, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
func Max(vs []uint64) uint64 {
	if len(vs) == 0 {
		return 0
	}
	m := vs[0]
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}
func Min(vs []uint64) uint64 {
	if len(vs) == 0 {
		return 0
	}
	m := vs[0]
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}
func Mean(vs []uint64) uint64 {
	if len(vs) == 0 {
		return 0
	}
	total := uint64(0)
	for _, v := range vs {
		total += v
	}
	return uint64(math.Round(float64(total) / float64(len(vs))))
}
func Median(vs []uint64) uint64 {
	if len(vs) == 0 {
		return 0
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i] < vs[j] })
	mNumber := len(vs) / 2

	if len(vs)%2 == 0 {
		return (vs[mNumber-1] + vs[mNumber]) / 2
	} else {
		return vs[mNumber]
	}
}
