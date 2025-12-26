package generate

import (
	"math/rand"
)

// create random string with first letter capitalized
func RandomStringCapitalized(maxLength int) string {
	if maxLength < 3 {
		maxLength = 3 // minimum 3 characters
	}

	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	// length random between 3 and maxLength
	length := rand.Intn(maxLength-2) + 3 // rand.Intn(max-min+1) + min

	s := make([]rune, length)
	s[0] = rune('A' + rand.Intn(26)) // first letter capitalized

	for i := 1; i < length; i++ {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}

// create phone number 10 digits start with 08 or 09
func RandomPhoneNumber() string {
	prefixes := []string{"08", "09"}
	prefix := prefixes[rand.Intn(len(prefixes))]

	number := make([]rune, 8)
	for i := 0; i < 8; i++ {
		number[i] = rune('0' + rand.Intn(10))
	}

	return prefix + string(number)
}
