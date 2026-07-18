package sliceutil

import (
	"reflect"
	"testing"
)

func TestUnique(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "nil", in: nil, want: []string{}},
		{name: "empty", in: []string{}, want: []string{}},
		{name: "unique", in: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "duplicates", in: []string{"a", "b", "a", "c", "b"}, want: []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("generic integers", func(t *testing.T) {
		if got, want := Unique([]int{1, 2, 1}), []int{1, 2}; !reflect.DeepEqual(got, want) {
			t.Errorf("Unique() = %v, want %v", got, want)
		}
	})
}
