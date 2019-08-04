package topic

import "testing"

func TestNewFilter(t *testing.T) {
	testcases := []struct {
		input   string
		success bool
	}{
		{"", false},
		{"/", true},
		{"sport/tennis/player1/#", true},
		{"#", true},
		{"sport/tennis#", false},
		{"sport/tennis/#/rankning", false},
		{"+", true},
		{"+/tennis/#", true},
		{"sport+", false},
		{"sport/+/player1", true},
		{"sport/player 1", true},
	}
	for _, c := range testcases {
		f := NewFilter(c.input)
		if c.success && f == nil {
			t.Errorf("Expected %s to succeed but didn't", c.input)
		}
		if !c.success && f != nil {
			t.Errorf("Expected %s to fail but didn't", c.input)
		}
	}
}

func TestMatch(t *testing.T) {
	testcases := []struct {
		filter      string
		matching    []string
		nonmatching []string
	}{
		{"sport/tennis/player1/#",
			[]string{
				"sport/tennis/player1",
				"sport/tennis/player1/ranking",
				"sport/tennis/player1/score/wimbledon",
			},
			[]string{},
		},
		{"sport/#",
			[]string{
				"sport",
			},
			[]string{},
		},
		{"#",
			[]string{
				"sport",
				"sport/tennis/player1",
				"sport/tennis/player1/ranking",
			},
			[]string{},
		},
		{"sport/tennis/+",
			[]string{
				"sport/tennis/player1",
				"sport/tennis/player2",
			},
			[]string{
				"sport/tennis/player1/ranking",
			},
		},
		{"sport/+",
			[]string{
				"sport/",
			},
			[]string{
				"sport",
			},
		},
	}

	for _, c := range testcases {
		f := NewFilter(c.filter)
		if f == nil {
			t.Fatalf("Unable to create filter from \"%v\"", c.filter)
		}
		for _, m := range c.matching {
			n := NewName(m)
			if n == nil {
				t.Fatalf("Unable to create name from \"%v\"", m)
			}
			if !f.Match(n) {
				t.Errorf("\"%v\" should match filter \"%v\"", m, c.filter)
			}
		}
		for _, m := range c.nonmatching {
			n := NewName(m)
			if n == nil {
				t.Fatalf("Unable to create name from \"%v\"", m)
			}
			if f.Match(n) {
				t.Errorf("\"%v\" should NOT match filter \"%v\"",
					m, c.filter)
			}
		}
	}
}
