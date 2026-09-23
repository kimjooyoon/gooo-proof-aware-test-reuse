package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProgramRejectsUnknownDeclaration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unknown.gooo")
	contents := "case strict\nprogram proof-aware\nnamespace test\nobligation smoke contract fixture go1.27.0 github-actions-ubuntu-latest\noutput \"ok\"\nunknwon value\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ParseProgram(path); err == nil {
		t.Fatal("unknown declaration was accepted")
	}
}
