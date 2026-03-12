package http

import (
	"encoding/base64"
	"encoding/json"
)

type tokenPayload struct {
	SID string `json:"sid"`
}

func split3(v string) []string {
	out := make([]string, 0, 3)
	cur := ""
	for _, ch := range v {
		if ch == '.' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(ch)
	}
	out = append(out, cur)
	return out
}

func decodePayload(raw string) (tokenPayload, error) {
	var p tokenPayload
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(decoded, &p); err != nil {
		return p, err
	}
	return p, nil
}
