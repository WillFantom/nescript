package nescript

import (
	"fmt"
	"strings"
)

// Formatter is a function that can convert a raw command (a string slice) into
// a single string. However, as different commands should be treated differently
// in this regard, the formatter dictates how this conversion happens.
type Formatter func([]string) string

var (
	defaultScriptFormatter Formatter = QuoteLastArgFormatter("\"", "\"")
	defaultCmdFormatter    Formatter = SpaceSepFormatter()
)

// SpaceSepFormatter joins all command and all arguments with a single space to
// create a single string value. This is perhaps the most basic formatter. For
// example, if a single argument contains spaces, the resulting string will
// present no differences between an argument that contains spaces, and the
// spaces used to separate arguments.
func SpaceSepFormatter() Formatter {
	return func(raw []string) string {
		if len(raw) <= 0 {
			return ""
		}
		return strings.Join(raw, " ")
	}
}

// QuoteIfSpaceFormatter acts similarly to SpaceSepFormatter, however if any
// argument contains a space, it will be wrapped in quotes.
func QuoteIfSpaceFormatter(openQuote, closeQuote string) Formatter {
	return func(raw []string) string {
		if len(raw) <= 0 {
			return ""
		}
		for idx, arg := range raw {
			if strings.Contains(arg, " ") {
				raw[idx] = fmt.Sprintf("%s%s%s", openQuote, raw[idx], closeQuote)
			}
		}
		return strings.Join(raw, " ")
	}
}

// QuoteLastArgFormatter joins all command and all arguments with a single space
// to create a string. However, it will wrap the final argument in quotes (using
// a the given quote chars).
func QuoteLastArgFormatter(openQuote, closeQuote string) Formatter {
	return func(raw []string) string {
		if len(raw) <= 0 {
			return ""
		}
		raw[len(raw)-1] = fmt.Sprintf("%s%s%s", openQuote, raw[len(raw)-1], closeQuote)
		return strings.Join(raw, " ")
	}
}
