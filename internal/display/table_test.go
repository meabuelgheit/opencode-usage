package display

import (
	"testing"
)

func TestNum(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{1234567890, "1,234,567,890"},
	}

	for _, tt := range tests {
		result := num(tt.input)
		if result != tt.expected {
			t.Errorf("num(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestCostStr(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "$0.00"},
		{0.01, "$0.01"},
		{1.00, "$1.00"},
		{1.5, "$1.50"},
		{123.456, "$123.46"},
		{0.005, "$0.01"}, // rounding
	}

	for _, tt := range tests {
		result := costStr(tt.input)
		if result != tt.expected {
			t.Errorf("costStr(%f) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestModelShort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gpt-4", "gpt-4"},
		{"claude-3-opus", "claude-3-opus"},
		{"openai/gpt-4", "openai/gpt-4"},
		{"openai/gpt-4/12345", "gpt-4/12345"},
		{"a/b/c/d/e", "d/e"},
	}

	for _, tt := range tests {
		result := modelShort(tt.input)
		if result != tt.expected {
			t.Errorf("modelShort(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
