package templateengine

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"valid date", time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), "Mar 15, 2024"},
		{"zero value", time.Time{}, ""},
		{"new year", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "Jan 01, 2025"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatDate(tt.input))
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"afternoon", time.Date(2024, 6, 20, 14, 30, 0, 0, time.UTC), "Jun 20, 2024 2:30 PM"},
		{"morning", time.Date(2024, 1, 5, 9, 15, 0, 0, time.UTC), "Jan 05, 2024 9:15 AM"},
		{"zero value", time.Time{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatDateTime(tt.input))
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "$0.00"},
		{"small amount", 5.99, "$5.99"},
		{"whole number", 100, "$100.00"},
		{"with thousands", 1234.56, "$1,234.56"},
		{"large amount", 10000, "$10,000.00"},
		{"negative", -42.50, "-$42.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatCurrency(tt.input))
		})
	}
}

func TestFormatMileage(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "0 mi"},
		{"small", 500, "500 mi"},
		{"with commas", 45230, "45,230 mi"},
		{"large", 150000, "150,000 mi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatMileage(tt.input))
		})
	}
}

func TestFormatMileagePtr(t *testing.T) {
	miles := 45230
	tests := []struct {
		name     string
		input    *int
		expected string
	}{
		{"nil", nil, "—"},
		{"with value", &miles, "45,230 mi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatMileagePtr(tt.input))
		})
	}
}

func TestCurrentYear(t *testing.T) {
	expected := fmt.Sprintf("%d", time.Now().Year())
	assert.Equal(t, expected, currentYear())
}

func TestSeq(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected []int
	}{
		{"zero", 0, []int{}},
		{"one", 1, []int{1}},
		{"five", 5, []int{1, 2, 3, 4, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := seq(tt.input)
			if tt.input == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	assert.Equal(t, 5, add(2, 3))
	assert.Equal(t, 0, add(-1, 1))
}

func TestSub(t *testing.T) {
	assert.Equal(t, 2, sub(5, 3))
	assert.Equal(t, -2, sub(3, 5))
}

func TestSafeHTML(t *testing.T) {
	result := safeHTML("<strong>bold</strong>")
	assert.Equal(t, "<strong>bold</strong>", string(result))
}

func TestBuildFuncMap(t *testing.T) {
	fm := buildFuncMap()

	expectedKeys := []string{
		"formatDate", "formatDateTime", "formatCurrency", "formatCurrencyPtr",
		"formatMileage", "formatMileagePtr",
		"toUpper", "toLower", "toTitle", "currentYear",
		"seq", "add", "sub", "safeHTML",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, fm, key, "FuncMap should contain %q", key)
	}
}
