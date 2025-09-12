package uidgen

import (
	"math/rand"
)

func UIDGen() string {
	digits := []rune("23456789")
	letters := []rune("ABCDEFGHJKLMNPQRTUVWXYZ")
	d1 := digits[rand.Intn(len(digits))]
	d2 := letters[rand.Intn(len(letters))]
	d3 := digits[rand.Intn(len(digits))]

	return string([]rune{d1, d2, d3})
}
