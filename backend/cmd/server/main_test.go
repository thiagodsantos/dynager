package main

import "testing"

func TestValidateLoopbackAddress(t *testing.T) {
	for _, address := range []string{"127.0.0.1:8080", "localhost:9000", "[::1]:8080"} {
		if err := validateLoopbackAddress(address); err != nil {
			t.Errorf("expected %s to be accepted: %v", address, err)
		}
	}
	for _, address := range []string{"0.0.0.0:8080", "192.168.1.1:8080", ":8080", "not-an-address"} {
		if err := validateLoopbackAddress(address); err == nil {
			t.Errorf("expected %s to be rejected", address)
		}
	}
}
