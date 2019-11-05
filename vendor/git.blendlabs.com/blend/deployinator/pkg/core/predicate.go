package core

// PredicateFunc is a function that returns bool and error
type PredicateFunc func() (bool, error)

// StringPredicateFunc is a predicate function that accepts one or more string argument
type StringPredicateFunc func(string, ...string) (bool, error)

// StringPredicate creates a PredicateFunc wrapper around a StringPredicateFunc
func StringPredicate(fn StringPredicateFunc, name string, namespace ...string) PredicateFunc {
	return func() (bool, error) {
		return fn(name, namespace...)
	}
}

// Not negates a predicate
func Not(p PredicateFunc) (bool, error) {
	ok, err := p()
	return !ok, err
}

// And evaluate logical AND of predicates (true if no predicates)
func And(predicates ...PredicateFunc) (bool, error) {
	for _, p := range predicates {
		ok, err := p()
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

// Or evaluate logical OR of predicates (false if no predicates)
func Or(predicates ...PredicateFunc) (bool, error) {
	for _, p := range predicates {
		ok, err := p()
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
