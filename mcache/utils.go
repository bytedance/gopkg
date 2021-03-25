package mcache

func isPowerOfTwo(x int) bool {
	return (x & (-x)) == x
}
