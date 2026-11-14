package main

import "testing"

func TestReceiverProblems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{name: "pointer receiver", source: "package p\n\ntype T struct{}\n\nfunc (t *T) M() {}\n"},
		{name: "generic pointer receiver", source: "package p\n\ntype T[V any] struct{}\n\nfunc (t *T[V]) M() {}\n"},
		{name: "plain function", source: "package p\n\nfunc F() {}\n"},
		{name: "value receiver", source: "package p\n\ntype T struct{}\n\nfunc (t T) M() {}\n", want: []string{"method M needs a pointer receiver"}},
		{name: "unnamed value receiver", source: "package p\n\ntype T struct{}\n\nfunc (T) M() {}\n", want: []string{"pointer receiver"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fileSet, file := parse(t, tt.source)

			assertProblems(t, receiverProblems(fileSet, file), tt.want)
		})
	}
}
