package util

import "fmt"

// PaddingZero pads the number n with zeros to the left to make it width digits long
func PaddingZero(n int, width int) string {
	format := "%0" + fmt.Sprint(width) + "d"
	return fmt.Sprintf(format, n)
}
