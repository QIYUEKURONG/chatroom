package main

import "testing"

// reference
// value
func TestMap(t *testing.T) {
	a := map[string]int{}
	b := a

	a["a"] = 1

	n, ok := b["a"]
	if !ok {
		t.Error("'a' not exist")
		return
	}

	if n != 1 {
		t.Errorf("n == %v", n)
	}
}
