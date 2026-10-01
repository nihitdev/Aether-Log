package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	r, err := Open(path, 4, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"aaaa", "bbbb", "cccc", "dddd"} {
		if _, err := r.Write([]byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	r.Close()
	for name, want := range map[string]string{path: "dddd", path + ".1": "cccc", path + ".2": "bbbb"} {
		got, err := os.ReadFile(name)
		if err != nil || string(got) != want {
			t.Fatalf("%s: %q %v", name, got, err)
		}
	}
	if _, err := os.Stat(path + ".3"); !os.IsNotExist(err) {
		t.Fatal("extra rotation")
	}
	if _, err := r.Write(nil); err == nil {
		t.Fatal("write after close")
	}
}
func TestExistingAndOversized(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	os.WriteFile(path, []byte("old"), 0600)
	r, err := Open(path, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Write([]byte("oversized")); err != nil {
		t.Fatal(err)
	}
	r.Close()
	data, _ := os.ReadFile(path)
	if string(data) != "oversized" {
		t.Fatal(string(data))
	}
}
