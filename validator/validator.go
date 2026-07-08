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
	"unicode"
)

var (
	ErrValidation = errors.New("validation failed")
)

type Validator struct {
	errors []error
}

func New() *Validator {
	return &Validator{
		errors: make([]error, 0),
	}
}

func (v *Validator) Required(value interface{}) *Validator {
	switch val := value.(type) {
	case string:
		if val == "" {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	case *string:
		if val == nil || *val == "" {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	case int:
		if val == 0 {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	case *int:
		if val == nil || *val == 0 {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	case int64:
		if val == 0 {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	case *int64:
		if val == nil || *val == 0 {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	case []interface{}:
		if len(val) == 0 {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	default:
		if value == nil {
			v.errors = append(v.errors, errors.New("field is required"))
		}
	}
	return v
}

func (v *Validator) Min(value string, min int) *Validator {
	if len(value) < min {
		v.errors = append(v.errors, fmt.Errorf("must be at least %d characters", min))
	}
	return v
}

func (v *Validator) Max(value string, max int) *Validator {
	if len(value) > max {
		v.errors = append(v.errors, fmt.Errorf("must be at most %d characters", max))
	}
	return v
}

func (v *Validator) Between(value string, min, max int) *Validator {
	if len(value) < min || len(value) > max {
		v.errors = append(v.errors, fmt.Errorf("must be between %d and %d characters", min, max))
	}
	return v
}

func (v *Validator) Email(value string) *Validator {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.errors = append(v.errors, errors.New("must be a valid email"))
	}
	return v
}

func (v *Validator) URL(value string) *Validator {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(value) {
		v.errors = append(v.errors, errors.New("must be a valid URL"))
	}
	return v
}

func (v *Validator) Numeric(value string) *Validator {
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(value) {
		v.errors = append(v.errors, errors.New("must be numeric"))
	}
	return v
}

func (v *Validator) Alpha(value string) *Validator {
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(value) {
		v.errors = append(v.errors, errors.New("must contain only letters"))
	}
	return v
}

func (v *Validator) AlphaNumeric(value string) *Validator {
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(value) {
		v.errors = append(v.errors, errors.New("must contain only letters and numbers"))
	}
	return v
}

func (v *Validator) Match(value, pattern string) *Validator {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.errors = append(v.errors, errors.New("format is invalid"))
	}
	return v
}

func (v *Validator) Equals(value, expected string) *Validator {
	if value != expected {
		v.errors = append(v.errors, errors.New("values do not match"))
	}
	return v
}

func (v *Validator) OneOf(value string, allowed []string) *Validator {
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.errors = append(v.errors, fmt.Errorf("must be one of: %v", allowed))
	return v
}

func (v *Validator) MinInt(value, min int64) *Validator {
	if value < min {
		v.errors = append(v.errors, fmt.Errorf("must be at least %d", min))
	}
	return v
}

func (v *Validator) MaxInt(value, max int64) *Validator {
	if value > max {
		v.errors = append(v.errors, fmt.Errorf("must be at most %d", max))
	}
	return v
}

func (v *Validator) Positive(value int64) *Validator {
	if value <= 0 {
		v.errors = append(v.errors, errors.New("must be positive"))
	}
	return v
}

func (v *Validator) Negative(value int64) *Validator {
	if value >= 0 {
		v.errors = append(v.errors, errors.New("must be negative"))
	}
	return v
}

func (v *Validator) HasUpperCase(value string) *Validator {
	hasUpper := false
	for _, r := range value {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		v.errors = append(v.errors, errors.New("must contain at least one uppercase letter"))
	}
	return v
}

func (v *Validator) HasLowerCase(value string) *Validator {
	hasLower := false
	for _, r := range value {
		if unicode.IsLower(r) {
			hasLower = true
			break
		}
	}
	if !hasLower {
		v.errors = append(v.errors, errors.New("must contain at least one lowercase letter"))
	}
	return v
}

func (v *Validator) HasDigit(value string) *Validator {
	hasDigit := false
	for _, r := range value {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		v.errors = append(v.errors, errors.New("must contain at least one digit"))
	}
	return v
}

func (v *Validator) HasSpecialChar(value string) *Validator {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	hasSpecial := false
	for _, r := range value {
		for _, c := range specialChars {
			if r == c {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	if !hasSpecial {
		v.errors = append(v.errors, errors.New("must contain at least one special character"))
	}
	return v
}

func (v *Validator) Custom(fn func() error) *Validator {
	if err := fn(); err != nil {
		v.errors = append(v.errors, err)
	}
	return v
}

func (v *Validator) Validate() error {
	if len(v.errors) > 0 {
		return fmt.Errorf("%w: %v", ErrValidation, v.errors)
	}
	return nil
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() []error {
	return v.errors
}
