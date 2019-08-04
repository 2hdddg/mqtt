package controlpacket

import "testing"

func TestNewClientId(t *testing.T) {
	testcases := []struct {
		id      string
		success bool
	}{
		{"", false},
		{"1", true},
		{"01234567890123456789012345678901234567890", false},
	}
	for _, c := range testcases {
		success := VerifyClientId(c.id)
		if success != c.success {
			t.Errorf("Expected %v but got %v", c.success, success)
		}
	}
}

