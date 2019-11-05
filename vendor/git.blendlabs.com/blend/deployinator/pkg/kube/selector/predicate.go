package selector

import (
	"fmt"
	"strings"
)

// Empty returns an empty selector.
func Empty() string {
	return ""
}

// Equals returns a selector that matches a label by a value.
func Equals(label, value string) string {
	return fmt.Sprintf("%s=%s", label, value)
}

// NotEquals represents an inequality comparator.
func NotEquals(label, value string) string {
	return fmt.Sprintf("%s!=%s", label, value)
}

// In represents set inclusion.
func In(label string, values ...string) string {
	return fmt.Sprintf("%s in (%s)", label, strings.Join(values, ","))
}

// NotIn represents set anti-inclusion.
func NotIn(label string, values ...string) string {
	return fmt.Sprintf("%s notin (%s)", label, strings.Join(values, ","))
}

// Has represents label existence inclusion filtering.
func Has(label string) string {
	return label
}

// NotHas represents label existence exclusion filtering.
func NotHas(label string) string {
	return fmt.Sprintf("!%s", label)
}
