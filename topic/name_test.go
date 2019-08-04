package topic

import "testing"

func TestNewName(t *testing.T) {
	testcases := []struct {
		input   string
		success bool
	}{
		{"", false},
		{"x y", true},
		{"/", true},
		{"sport/tennis/player1", true},
		{"$xyz/abc", false},
	}
	for _, c := range testcases {
		f := NewName(c.input)
		if c.success && f == nil {
			t.Errorf("Expected %s to succeed but didn't", c.input)
		}
		if !c.success && f != nil {
			t.Errorf("Expected %s to fail but didn't", c.input)
		}
	}
}
