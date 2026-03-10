package valueobject

import (
	"fmt"
	"strings"
)

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeFile  ContentType = "file"
)

func ParseContentType(v string) (ContentType, error) {
	switch ContentType(strings.ToLower(v)) {
	case ContentTypeText, ContentTypeImage, ContentTypeFile:
		return ContentType(strings.ToLower(v)), nil
	default:
		return "", fmt.Errorf("unsupported content type: %s", v)
	}
}

type Cursor struct {
	Value string
	Limit int
}

func NewCursor(value string, limit int) Cursor {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return Cursor{Value: value, Limit: limit}
}
