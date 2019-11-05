package selector

import "strings"

// And represents a logical group of selectors.
func And(predicates ...string) string {
	return strings.Join(predicates, ",")
}
