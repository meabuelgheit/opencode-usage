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
		{1000, "1.00K"},
		{1234567, "1.23M"},
		{1500, "1.50K"},
		{10240, "10.24K"},
		{999999, "1000.00K"},
		{1000000, "1.00M"},
		{15000000, "15.00M"},
		{1000000000, "1.00B"},
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

func TestPctStr(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0.0%"},
		{50.0, "50.0%"},
		{95.234, "95.2%"},
		{100.0, "100.0%"},
		{0.05, "0.1%"}, // rounded up
	}

	for _, tt := range tests {
		result := pctStr(tt.input)
		if result != tt.expected {
			t.Errorf("pctStr(%f) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}
