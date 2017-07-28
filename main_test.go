package main

import (
	"testing"
)

func TestParseThreshold(t *testing.T) {
	for name, c := range map[string]struct {
		input     string
		output    int64
		shouldErr bool
	}{
		"k": {
			input:  "7k",
			output: 7 * K,
		},
		"K": {
			input:  "7K",
			output: 7 * K,
		},
		"m": {
			input:  "7m",
			output: 7 * M,
		},
		"g": {
			input:  "7g",
			output: 7 * G,
		},
		"noMultiplier": {
			input:  "7",
			output: 7,
		},
		"invalidMultiplier": {
			input:     "7kb",
			output:    -1,
			shouldErr: true,
		},
		"invalidNumber": {
			input:     "7bk",
			output:    -1,
			shouldErr: true,
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			got, err := parseThreshold(c.input)
			if c.shouldErr && err == nil {
				t.Errorf("should get error, but not")
			}
			if !c.shouldErr && err != nil {
				t.Errorf("got unexpected error: %s", err)
			}
			if got != c.output {
				t.Errorf("expect %d, but got %d", c.output, got)
			}
		})
	}
}
