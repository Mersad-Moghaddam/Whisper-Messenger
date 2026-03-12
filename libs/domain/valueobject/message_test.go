package valueobject

import "testing"

func TestParseContentType(t *testing.T) {
	if _, err := ParseContentType("text"); err != nil {
		t.Fatalf("expected text type to parse: %v", err)
	}
	if _, err := ParseContentType("unsupported"); err == nil {
		t.Fatalf("expected unsupported type error")
	}
}

func TestNewCursor_DefaultLimit(t *testing.T) {
	c := NewCursor("", 0)
	if c.Limit != 50 {
		t.Fatalf("expected default limit 50, got %d", c.Limit)
	}
}
