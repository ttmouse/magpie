package settings

import (
	"encoding/json"
	"os"
	"testing"
)

// A file that never chose the Dock mode gets the default (#1381): the main
// window open puts magpie in the Dock and Cmd+Tab as a regular app.
func TestDockWindowDefaultsWhenUnset(t *testing.T) {
	for _, body := range []string{
		`{}`,
		`{"theme":"dark","lang":"zh"}`,
		"\ufeff{}",
	} {
		t.Run(body, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			if err := os.MkdirAll(Dir(), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(Path(), []byte(body), 0o640); err != nil {
				t.Fatal(err)
			}
			if !Load().DockWindow {
				t.Fatalf("a file that never said (%q) must default DockWindow to true", body)
			}
		})
	}
}

// A first run has no file at all; the default holds there too, and the
// first save writes the key so the choice can stick either way.
func TestDockWindowDefaultsOnFirstRun(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if !Load().DockWindow {
		t.Fatalf("a magpie with no settings file must default DockWindow to true")
	}
	s := Load()
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(Path())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if json.Unmarshal(b, &raw) != nil {
		t.Fatalf("the first save wrote a file that does not parse:\n%s", b)
	}
	if _, ok := raw["dockWindow"]; !ok {
		t.Fatalf("the first save must write the key, got:\n%s", b)
	}
}

// A choice that is on the file is kept, off as much as on.
func TestDockWindowKeepsExplicitChoice(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{`{"dockWindow":false}`, false},
		{`{"dockWindow":true}`, true},
	} {
		t.Run(tc.body, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			if err := os.MkdirAll(Dir(), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(Path(), []byte(tc.body), 0o640); err != nil {
				t.Fatal(err)
			}
			if got := Load().DockWindow; got != tc.want {
				t.Fatalf("%s: DockWindow = %v, want %v", tc.body, got, tc.want)
			}
		})
	}
}

// Turning the Dock off keeps saying so on the file: a false that Save
// dropped would read as "never chose" and come back as the default.
func TestSaveWritesDockWindowOff(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		t.Fatal(err)
	}
	s := Load()
	if !s.DockWindow {
		t.Fatalf("a fresh file defaults DockWindow to true")
	}
	s.DockWindow = false
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(Path())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if json.Unmarshal(b, &raw) != nil {
		t.Fatalf("the saved file does not parse:\n%s", b)
	}
	v, ok := raw["dockWindow"]
	if !ok || v != false {
		t.Fatalf("Save must write the key even when false, got:\n%s", b)
	}
	if Load().DockWindow {
		t.Fatalf("the choice written off did not load back off")
	}
}
