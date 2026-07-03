package main

import "testing"

func TestMainFunc(t *testing.T) {
	main()
}

func TestRun(t *testing.T) {
	if Run() != 0 {
		t.Error("Run() should return 0")
	}
}
