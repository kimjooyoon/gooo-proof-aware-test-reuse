package engine

import "testing"

func TestRejectDuplicateJSONKeysRejectsNestedKeys(t *testing.T) {
	if err := rejectDuplicateJSONKeys([]byte(`{"proof":{"digest":"first","digest":"second"}}`)); err == nil {
		t.Fatal("expected duplicate JSON key to be rejected")
	}
}

func TestRejectDuplicateJSONKeysAcceptsUniqueKeys(t *testing.T) {
	if err := rejectDuplicateJSONKeys([]byte(`{"proof":{"digest":"first","status":"PASS"}}`)); err != nil {
		t.Fatalf("expected unique JSON keys to be accepted: %v", err)
	}
}
