package jsonlen

// Int64 returns the encoded length of d as a JSON Number.
// Int64(0) == 1, Int64(-50) == 3, Int64(7654321) = 7
func Int64(d int64) int {
	sign := 0
	if d < 0 {
		sign++
		d *= -1
	}
	return sign + Uint64(uint64(d))
}

// Uint64 returns the encoded length of d as a JSON Number.
// Uint64(0) == 1, Int64(50) == 2, Int64(7654321) = 7
func Uint64(d uint64) int {
	l := 1
	for pow := uint64(10); d/pow > 0; pow *= 10 {
		l++
	}
	return l
}
