package main

import "testing"

var pingTests = []struct {
	input string
	err   error
}{
	{"127.0.0.1", nil},
}

func TestPingTest(t *testing.T) {

	for _, test := range pingTests {
		if actual := pingTest(test.input, "10"); actual != nil {
			t.Errorf("pingTest(%q) = %q, expected nil",
				test.input, actual)
		}
	}

}

func TestGetEnvVar(t *testing.T) {

	for i := 1; i < 5; i++ {
		if GetEnvVar("PING_DESTINATION", "10.10.10.10") == "" {
			t.Errorf("GetEnvVar(10.10.10.10), expected a non empty string")
		}
	}

}
