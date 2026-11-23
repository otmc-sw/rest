/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package schema

import "strings"

// Snake converts CamelCase to snake_case: "RoleID" -> "role_id".
func Snake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			// Insert underscore before an uppercase that follows a lowercase/digit,
			// or before the last capital of an acronym ("HTTPServer" -> "http_server").
			if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') ||
				(prev >= 'A' && prev <= 'Z' && i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z') {
				b.WriteByte('_')
			}
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Plural returns a naive English plural: "user" -> "users", "category" -> "categories".
func Plural(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	switch {
	case strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"),
		strings.HasSuffix(lower, "ch"), strings.HasSuffix(lower, "sh"):
		return s + "es"
	case strings.HasSuffix(lower, "y") && len(lower) > 1 && !isVowel(lower[len(lower)-2]):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(lower, "f"):
		return s[:len(s)-1] + "ves"
	default:
		return s + "s"
	}
}

func isVowel(c byte) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
