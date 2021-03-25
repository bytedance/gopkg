// +build !amd64

package mcache

func bsr(x int) int {
	r := 0
	for x != 0 {
		x = x >> 1
		r += 1
	}
	return r - 1
}
