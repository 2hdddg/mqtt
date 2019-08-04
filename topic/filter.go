package topic

import "strings"

type wildmulti int
type wildsingle int

type Filter struct {
	parts []interface{}
}

func NewFilter(s string) *Filter {
	if s == "" {
		return nil
	}

	splitted := strings.Split(s, SEP)
	parts := make([]interface{}, 0, len(splitted))

	for i, n := range splitted {
		if n == WILDCARD_MULTI {
			if i < (len(splitted) - 1) {
				return nil
			}
			parts = append(parts, wildmulti(0))
			continue
		}
		if n == WILDCARD_SINGLE {
			parts = append(parts, wildsingle(0))
			continue
		}
		if strings.Contains(n, WILDCARD_MULTI) ||
			strings.Contains(n, WILDCARD_SINGLE) {
			return nil
		}
		parts = append(parts, n)
	}

	return &Filter{parts: parts}
}

func (f *Filter) Match(n *Name) bool {
	if f == nil || n == nil {
		return false
	}

	ni := 0
	for _, fx := range f.parts {
		switch v := fx.(type) {
		case wildmulti:
			return true
		case wildsingle:
			if ni >= len(n.parts) {
				return false
			}
			ni++
		case string:
			if ni >= len(n.parts) {
				return false
			}
			if n.parts[ni] != v {
				return false
			}
			ni++
		}
	}

	return ni == len(n.parts)
}

