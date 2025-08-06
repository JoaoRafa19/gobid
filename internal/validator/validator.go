package validator

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator interface {
	Valid(ctx context.Context) Evaluator
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Evaluator map[string]string

func (e *Evaluator) addFieldError(key, message string) {
	if *e == nil {
		*e = make(map[string]string)
	}

	if _, ok := (*e)[key]; !ok {
		(*e)[key] = message
	}
}

func (e *Evaluator) CheckField(ok bool, key, message string) {
	if !ok {
		e.addFieldError(key, message)
	}
}

func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

func MaxChar(s string, max int) bool {
	return utf8.RuneCountInString(s) <= max
}

func MinChar(s string, min int) bool {
	return utf8.RuneCountInString(s) >= min
}

func Matches(s string, regex *regexp.Regexp) bool {
	return regex.MatchString(s)
}
