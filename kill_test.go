package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestVictimsSort(t *testing.T) {
	for name, c := range map[string]struct {
		input, expect victims
	}{
		"oomAdj": {
			input: victims{
				&task{3, "3", 789, 2},
				&task{2, "2", 456, 12},
				&task{1, "1", 123, 10},
			},
			expect: victims{
				&task{2, "2", 456, 12},
				&task{1, "1", 123, 10},
				&task{3, "3", 789, 2},
			},
		},

		"rss": {
			input: victims{
				&task{3, "3", 123, 7},
				&task{2, "2", 456, 7},
				&task{1, "1", 789, 7},
			},
			expect: victims{
				&task{1, "1", 789, 7},
				&task{2, "2", 456, 7},
				&task{3, "3", 123, 7},
			},
		},

		"pid": {
			input: victims{
				&task{1, "1", 123, 7},
				&task{2, "2", 123, 7},
				&task{3, "3", 123, 7},
			},
			expect: victims{
				&task{3, "3", 123, 7},
				&task{2, "2", 123, 7},
				&task{1, "1", 123, 7},
			},
		},

		"oomAdj+rss+pid": {
			input: victims{
				&task{3, "3", 789, 2},
				&task{2, "2", 456, 12},
				&task{5, "2", 456, 12},
				&task{1, "1", 123, 10},
				&task{4, "1", 234, 10},
			},
			expect: victims{
				&task{5, "2", 456, 12},
				&task{2, "2", 456, 12},
				&task{4, "1", 234, 10},
				&task{1, "1", 123, 10},
				&task{3, "3", 789, 2},
			},
		},
	} {
		c := c
		t.Run(name, func(t *testing.T) {
			got := c.input
			if sort.Sort(got); !reflect.DeepEqual(got, c.expect) {
				t.Errorf("expected:\n%v, got:\n%v\n", c.expect, got)
			}
		})
	}
}
