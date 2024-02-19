package shapes

import (
	"strconv"
	"strings"
)

var supdigits = []rune(`⁰¹²³⁴⁵⁶⁷⁸⁹`)

func supInts(a []int) (retVal string) {
	b := new(strings.Builder)
	b.WriteRune('⁽')
	for i, num := range a {
		// Convert the integer to string
		str := strconv.Itoa(num)
		// Loop through each character in the string and convert it to superscript
		for _, char := range str {
			offset := int(char) - int('0')
			sup := supdigits[offset]
			b.WriteRune(sup)
		}
		if i < len(a)-1 {
			b.WriteRune(' ')
		}
	}
	b.WriteRune('⁾')
	return b.String()
}

func supInt(a int) (retVal string) {
	str := strconv.Itoa(a)
	for _, r := range str {
		offset := int(r) - int('0')
		retVal += string(supdigits[offset])
	}
	return retVal
}
