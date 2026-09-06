package validation

import "testing"

func TestRequired(t *testing.T) {
	if Required("  ") {
		t.Fatal("whitespace must not be required input")
	}
	if !Required("BandRoom") {
		t.Fatal("non-empty input should be accepted")
	}
}

func TestEmail(t *testing.T) {
	for _, input := range []string{"owner@example.com", "customer@demo.test"} {
		if !Email(input) {
			t.Fatalf("expected valid email: %s", input)
		}
	}
	for _, input := range []string{"", "@example.com", "owner@"} {
		if Email(input) {
			t.Fatalf("expected invalid email: %s", input)
		}
	}
}
