package util

// CountDigits returns the number of digits in an integer
func CountDigits(n int) int {
	if n == 0 {
		return 1
	}

	count := 0
	if n < 0 {
		n = -n
		count++
	}

	for n != 0 {
		n /= 10
		count++
	}

	return count
}
