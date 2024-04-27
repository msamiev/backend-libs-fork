package primitives

import (
	"strconv"
	"testing"
)

func TestBuckets(t *testing.T) {
	cases := []struct {
		in   []int
		size int
		out  [][]int
	}{
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 3, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {0}}},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 2, [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 0}}},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 1, [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {0}}},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 4, [][]int{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 0}}},
		{[]int{}, 3, [][]int{}},
		{nil, -1, [][]int{}},
	}

	for i, c := range cases {
		c := c // pin

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			buckets := Buckets(c.in, c.size)

			if len(buckets) != len(c.out) {
				t.Errorf("Buckets() = %v, want %v", buckets, c.out)
			}

			for i, bucket := range buckets {
				for j, item := range bucket {
					if item != c.out[i][j] {
						t.Errorf("Buckets(%d:%d) = %v, want %v", i, j, buckets, c.out)
					}
				}
			}
		})
	}
}
