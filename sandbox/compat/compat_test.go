package compat_test

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/digitalocean/godo/sandbox/compat"
	"github.com/digitalocean/godo/sandbox/compat/daytona"
	"github.com/digitalocean/godo/sandbox/compat/daytona/types"
	"github.com/digitalocean/godo/sandbox/compat/modal"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

func TestErrorAliasesMatchTheLeafPackage(t *testing.T) {
	err := unsupported.New("Daytona", "Sandbox.Archive", "a sandbox auto-archive API")
	var nie *compat.NotImplementedError
	if !errors.As(err, &nie) || !errors.Is(err, compat.ErrNotImplemented) {
		t.Fatalf("%v does not match the compat names", err)
	}
	want := "Daytona.Sandbox.Archive is not implemented on DigitalOcean Sandboxes yet: needs a sandbox auto-archive API"
	if err.Error() != want {
		t.Errorf("message = %q, want %q", err, want)
	}
}

func TestConstructors(t *testing.T) {
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "do-token")
	var d *compat.DaytonaClient
	d, err := compat.NewDaytonaClient(&types.DaytonaConfig{APIKey: "k"})
	if err != nil {
		t.Fatal(err)
	}
	var _ *daytona.Client = d
	var m *compat.ModalClient
	m, err = compat.NewModalClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	var _ *modal.Client = m
}

func TestGodoDoesNotImportTheCompatPackages(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "github.com/digitalocean/godo", "github.com/digitalocean/godo/sandbox").CombinedOutput()
	if err != nil {
		t.Skipf("go list failed: %v: %s", err, out)
	}
	if strings.Contains(string(out), "sandbox/compat") {
		t.Errorf("godo or godo/sandbox depends on a compat package:\n%s", out)
	}
}
