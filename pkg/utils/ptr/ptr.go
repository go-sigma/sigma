// Package ptr ...
package ptr

// Of returns pointer to value.
func Of[T any](v T) *T {
	return &v
}

// To returns the value of the pointer passed in or the default value if the pointer is nil.
func To[T any](v *T) T {
	var zero T
	if v == nil {
		return zero
	}
	return *v
}

// ToDef returns the value of the int pointer passed in or default value if the pointer is nil.
func ToDef[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}
