package topic

import "strings"

const (
	SEP             = "/"
	WILDCARD_MULTI  = "#"
	WILDCARD_SINGLE = "+"
)

type Name struct {
	parts []string
}

func NewName(s string) *Name {
	if s == "" {
		return nil
	}

	if strings.HasPrefix(s, "$") {
		return nil
	}

	parts := strings.Split(s, SEP)
	return &Name{parts: parts}
}
