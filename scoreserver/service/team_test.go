package service

import (
	"testing"
)

func TestValidateContryCode(t *testing.T) {
	a := newApp(t)
	appStruct := a.(*app)

	testCases := []struct {
		input    string
		output   string
		hasError bool
	}{
		{"JPN", "JPN", false},
		{"JP", "JPN", false},
		{"", "", false},
		{"JAPAN", "", true},
	}

	for _, c := range testCases {
		code, err := appStruct.validateCountryCode(c.input)
		if c.hasError != (err != nil) {
			t.Errorf("case %v, err: %v", c, err)
		}
		if code != c.output {
			t.Errorf("expect: %v, actual: %v", c.output, code)
		}
	}
}
