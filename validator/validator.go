/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrValidation = errors.New("validation failed")

	reEmail      = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	reURL        = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	reNumeric    = regexp.MustCompile(`^[0-9]+$`)
	reAlpha      = regexp.MustCompile(`^[a-zA-Z]+$`)
	reAlphaNum   = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

type Validator struct {
	errors []error
	field  string
}

func New() *Validator {
	return &Validator{errors: make([]error, 0)}
}

func (v *Validator) Field(name string) *Validator {
	v.field = name
	return v
}

func (v *Validator) addErr(msg string) {
	if v.field != "" {
		msg = fmt.Sprintf("%s: %s", v.field, msg)
	}
	v.errors = append(v.errors, errors.New(msg))
}

func derefString(val any) (string, bool) {
	switch s := val.(type) {
	case string:
		return s, true
	case *string:
		if s == nil {
			return "", false
		}
		return *s, true
	default:
		return "", false
	}
}

func derefInt(val any) (int64, bool) {
	switch n := val.(type) {
	case int:
		return int64(n), true
	case *int:
		if n == nil {
			return 0, false
		}
		return int64(*n), true
	case int64:
		return n, true
	case *int64:
		if n == nil {
			return 0, false
		}
		return *n, true
	default:
		return 0, false
	}
}

func matchRegex(val any, re *regexp.Regexp, errMsg string, v *Validator) *Validator {
	s, ok := derefString(val)
	if !ok {
		return v
	}
	if !re.MatchString(s) {
		v.addErr(errMsg)
	}
	return v
}

func checkStringFn(val any, fn func(string) bool, errMsg string, v *Validator) *Validator {
	s, ok := derefString(val)
	if !ok {
		return v
	}
	if !fn(s) {
		v.addErr(errMsg)
	}
	return v
}

func (v *Validator) Required(value any) *Validator {
	var empty bool
	switch val := value.(type) {
	case string:
		empty = strings.TrimSpace(val) == ""
	case *string:
		empty = val == nil || strings.TrimSpace(*val) == ""
	case int:
		empty = val == 0
	case *int:
		empty = val == nil || *val == 0
	case int64:
		empty = val == 0
	case *int64:
		empty = val == nil || *val == 0
	case []any:
		empty = len(val) == 0
	default:
		empty = value == nil
	}
	if empty {
		v.addErr("field is required")
	}
	return v
}

func (v *Validator) Min(value any, min int) *Validator {
	return checkStringFn(value, func(s string) bool { return len(s) >= min },
		fmt.Sprintf("must be at least %d characters", min), v)
}

func (v *Validator) Max(value any, max int) *Validator {
	return checkStringFn(value, func(s string) bool { return len(s) <= max },
		fmt.Sprintf("must be at most %d characters", max), v)
}

func (v *Validator) Between(value any, min, max int) *Validator {
	return checkStringFn(value, func(s string) bool { return len(s) >= min && len(s) <= max },
		fmt.Sprintf("must be between %d and %d characters", min, max), v)
}

func (v *Validator) Email(value any) *Validator {
	return matchRegex(value, reEmail, "must be a valid email", v)
}

func (v *Validator) URL(value any) *Validator {
	return matchRegex(value, reURL, "must be a valid URL", v)
}

func (v *Validator) Numeric(value any) *Validator {
	return matchRegex(value, reNumeric, "must be numeric", v)
}

func (v *Validator) Alpha(value any) *Validator {
	return matchRegex(value, reAlpha, "must contain only letters", v)
}

func (v *Validator) AlphaNumeric(value any) *Validator {
	return matchRegex(value, reAlphaNum, "must contain only letters and numbers", v)
}

func (v *Validator) Match(value any, pattern string) *Validator {
	s, ok := derefString(value)
	if !ok {
		return v
	}
	matched, err := regexp.MatchString(pattern, s)
	if err != nil || !matched {
		v.addErr("format is invalid")
	}
	return v
}

func (v *Validator) Equals(value any, expected string) *Validator {
	return checkStringFn(value, func(s string) bool { return s == expected },
		"values do not match", v)
}

func (v *Validator) OneOf(value any, allowed []string) *Validator {
	s, ok := derefString(value)
	if !ok {
		return v
	}
	for _, a := range allowed {
		if s == a {
			return v
		}
	}
	v.addErr(fmt.Sprintf("must be one of: %v", allowed))
	return v
}

func (v *Validator) HasUpperCase(value any) *Validator {
	return checkStringFn(value, func(s string) bool {
		for _, r := range s {
			if unicode.IsUpper(r) {
				return true
			}
		}
		return false
	}, "must contain at least one uppercase letter", v)
}

func (v *Validator) HasLowerCase(value any) *Validator {
	return checkStringFn(value, func(s string) bool {
		for _, r := range s {
			if unicode.IsLower(r) {
				return true
			}
		}
		return false
	}, "must contain at least one lowercase letter", v)
}

func (v *Validator) HasDigit(value any) *Validator {
	return checkStringFn(value, func(s string) bool {
		for _, r := range s {
			if unicode.IsDigit(r) {
				return true
			}
		}
		return false
	}, "must contain at least one digit", v)
}

func (v *Validator) HasSpecialChar(value any) *Validator {
	return checkStringFn(value, func(s string) bool {
		return strings.ContainsAny(s, specialChars)
	}, "must contain at least one special character", v)
}

func (v *Validator) MinInt(value any, min int64) *Validator {
	n, ok := derefInt(value)
	if !ok {
		return v
	}
	if n < min {
		v.addErr(fmt.Sprintf("must be at least %d", min))
	}
	return v
}

func (v *Validator) MaxInt(value any, max int64) *Validator {
	n, ok := derefInt(value)
	if !ok {
		return v
	}
	if n > max {
		v.addErr(fmt.Sprintf("must be at most %d", max))
	}
	return v
}

func (v *Validator) Positive(value any) *Validator {
	n, ok := derefInt(value)
	if !ok {
		return v
	}
	if n <= 0 {
		v.addErr("must be positive")
	}
	return v
}

func (v *Validator) Negative(value any) *Validator {
	n, ok := derefInt(value)
	if !ok {
		return v
	}
	if n >= 0 {
		v.addErr("must be negative")
	}
	return v
}

func (v *Validator) Custom(fn func() error) *Validator {
	if err := fn(); err != nil {
		v.addErr(err.Error())
	}
	return v
}

func (v *Validator) Validate() error {
	if len(v.errors) > 0 {
		return fmt.Errorf("%w: %v", ErrValidation, v.errors)
	}
	return nil
}

func (v *Validator) Process() error { return v.Validate() }

func (v *Validator) HasErrors() bool { return len(v.errors) > 0 }
func (v *Validator) Errors() []error { return v.errors }
