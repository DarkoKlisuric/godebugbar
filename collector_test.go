package godebugbar

import (
	"net/http"
	"testing"
)

func TestStatusColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   BadgeColor
	}{
		{name: "100 informational is default", status: http.StatusContinue, want: ColorDefault},
		{name: "200 ok is green", status: http.StatusOK, want: ColorGreen},
		{name: "201 created is green", status: http.StatusCreated, want: ColorGreen},
		{name: "302 redirect is default", status: http.StatusFound, want: ColorDefault},
		{name: "400 bad request is yellow", status: http.StatusBadRequest, want: ColorYellow},
		{name: "404 not found is yellow", status: http.StatusNotFound, want: ColorYellow},
		{name: "500 internal error is red", status: http.StatusInternalServerError, want: ColorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := StatusColor(tt.status); got != tt.want {
				t.Errorf("StatusColor(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
