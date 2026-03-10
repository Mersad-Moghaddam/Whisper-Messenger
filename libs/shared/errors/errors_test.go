package errors_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "whisper/libs/shared/errors"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "validation", err: apperrors.New(apperrors.KindValidation, "bad_input", "invalid", nil), want: http.StatusBadRequest},
		{name: "unauthorized", err: apperrors.New(apperrors.KindUnauthorized, "unauth", "unauthorized", nil), want: http.StatusUnauthorized},
		{name: "internal", err: apperrors.New(apperrors.KindInternal, "internal", "internal", nil), want: http.StatusInternalServerError},
		{name: "unknown", err: errors.New("oops"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apperrors.HTTPStatus(tt.err); got != tt.want {
				t.Fatalf("HTTPStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}
