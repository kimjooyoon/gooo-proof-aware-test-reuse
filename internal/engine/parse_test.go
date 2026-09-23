package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProgramPreservesHashInsideQuotedOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hash-output.gooo")
	contents := "case hash-output\nprogram proof-aware\nnamespace test\nobligation smoke contract fixture go1.27.0 github-actions-ubuntu-latest\noutput \"hello # from Gooo\"\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	program, err := ParseProgram(path)
	if err != nil {
		t.Fatal(err)
	}
	if program.Output != "hello # from Gooo" {
		t.Fatalf("output = %q", program.Output)
	}
}
