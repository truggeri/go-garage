package templateengine

import (
	"fmt"
	"html/template"
	"strings"
	"time"
	"unicode"
)

// buildFuncMap returns a FuncMap containing all helper functions available in templates.
func buildFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate":        formatDate,
		"formatDateTime":    formatDateTime,
		"formatCurrency":    formatCurrency,
		"formatCurrencyPtr": formatCurrencyPtr,
		"formatMileage":     formatMileage,
		"formatMileagePtr":  formatMileagePtr,
		"toUpper":           strings.ToUpper,
		"toLower":           strings.ToLower,
		"toTitle":           titleCase,
		"currentYear":       currentYear,
		"seq":               seq,
		"add":               add,
		"sub":               sub,
		"safeHTML":          safeHTML,
	}
}

// formatDate formats a time.Time as "Jan 02, 2006".
func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 02, 2006")
}

// formatDateTime formats a time.Time as "Jan 02, 2006 3:04 PM".
func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 02, 2006 3:04 PM")
}

// formatCurrency formats a float as USD currency (e.g. "$1,234.56").
func formatCurrency(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	whole := int64(amount)
	cents := int64((amount - float64(whole) + 0.005) * 100)
	if cents >= 100 {
		whole++
		cents = 0
	}

	wholeStr := fmt.Sprintf("%d", whole)
	if len(wholeStr) > 3 {
		var result []byte
		for i, c := range wholeStr {
			pos := len(wholeStr) - i
			if pos%3 == 0 && i > 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		wholeStr = string(result)
	}

	if negative {
		return fmt.Sprintf("-$%s.%02d", wholeStr, cents)
	}
	return fmt.Sprintf("$%s.%02d", wholeStr, cents)
}

// formatCurrencyPtr formats a *float64 as USD currency. Returns "—" if nil.
func formatCurrencyPtr(amount *float64) string {
	if amount == nil {
		return "—"
	}
	return formatCurrency(*amount)
}

// formatMileagePtr formats a *int mileage with commas (e.g. "45,230 mi"). Returns "—" if nil.
func formatMileagePtr(miles *int) string {
	if miles == nil {
		return "—"
	}
	return formatMileage(*miles)
}

// formatMileage formats an integer mileage with commas (e.g. "45,230 mi").
func formatMileage(miles int) string {
	s := fmt.Sprintf("%d", miles)
	if len(s) <= 3 {
		return s + " mi"
	}

	var result []byte
	for i, c := range s {
		pos := len(s) - i
		if pos%3 == 0 && i > 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result) + " mi"
}

// currentYear returns the current year as a string.
func currentYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

// seq generates a slice of integers from 1 to n, useful for pagination.
func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i + 1
	}
	return s
}

// add returns the sum of two integers.
func add(a, b int) int {
	return a + b
}

// sub returns the difference of two integers.
func sub(a, b int) int {
	return a - b
}

// safeHTML marks a string as safe HTML that should not be escaped.
// Use with caution – only for trusted content.
func safeHTML(s string) template.HTML {
	return template.HTML(s) //nolint:gosec
}

// titleCase capitalises the first letter of each word in s.
func titleCase(s string) string {
	prev := ' '
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(prev) {
			prev = r
			return unicode.ToTitle(r)
		}
		prev = r
		return r
	}, s)
}
