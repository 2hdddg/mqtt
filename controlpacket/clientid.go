package controlpacket

import (

	"unicode"
)

func VerifyClientId(id string) bool {
	if len(id) > 23 || len(id) == 0 {
		return false
	}
	for _, r := range id {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r)) {
			return false
		}
	}
	return true
}
