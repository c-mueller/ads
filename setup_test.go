package ads

import (
	"github.com/mholt/caddy"
	"testing"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `ads`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}

	c = caddy.NewTestController("dns", `ads more`)
	if err := setup(c); err == nil {
		t.Fatalf("Expected errors, but got: %v", err)
	}
}
