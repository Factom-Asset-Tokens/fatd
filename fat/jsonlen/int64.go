package jsonlen

func Int64(d int64) int {
	sign := 0
	if d < 0 {
		sign++
		d *= -1
	}
	return sign + Uint64(uint64(d))
}

func Uint64(d uint64) int {
	l := 1
	for pow := uint64(10); d/pow > 0; pow *= 10 {
		l++
	}
	return l
}
