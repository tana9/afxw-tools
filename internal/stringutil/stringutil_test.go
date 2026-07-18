package stringutil

import (
	"reflect"
	"strings"
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{`C:\Windows`, `C:\Users`, `C:\Temp`},
			expected: []string{`C:\Windows`, `C:\Users`, `C:\Temp`},
		},
		{
			name:     "with duplicates",
			input:    []string{`C:\Windows`, `C:\Users`, `C:\Windows`, `C:\Temp`},
			expected: []string{`C:\Windows`, `C:\Users`, `C:\Temp`},
		},
		{
			name:     "all duplicates",
			input:    []string{`C:\Windows`, `C:\Windows`, `C:\Windows`},
			expected: []string{`C:\Windows`},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{`C:\Windows`},
			expected: []string{`C:\Windows`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicates(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRemoveDuplicatesBy(t *testing.T) {
	input := []string{"A", "a", "B"}
	result := RemoveDuplicatesBy(input, func(s string) string {
		return strings.ToLower(s)
	})
	expected := []string{"A", "B"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
