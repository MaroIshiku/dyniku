package server

import "testing"

func TestNormalizeJSONRequiresSettingsArray(t *testing.T) {
	t.Parallel()

	_, err := normalizeJSON([]byte(`{"settings":{}}`))
	if err == nil {
		t.Fatal("expected object settings to be rejected")
	}
}

func TestNormalizeJSONAddsEmptySettingsArray(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeJSON([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(normalized) != "{\n  \"settings\": []\n}" {
		t.Fatalf("unexpected normalized JSON:\n%s", normalized)
	}
}

func TestNormalizeJSONAcceptsUTF8BOM(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeJSON([]byte("\xef\xbb\xbf" + `{"settings":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(normalized) != "{\n  \"settings\": []\n}" {
		t.Fatalf("unexpected normalized JSON:\n%s", normalized)
	}
}
