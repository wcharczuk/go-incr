package testutil

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

// Error is a test helper to verify that an error is set.
func Error(t *testing.T, err error, message ...any) {
	t.Helper()
	if isNil(err) {
		fatalf(t, "expected error to not be <nil>", nil, message)
	}
}

// NoError is a test helper to verify that an error is nil.
//
// It is useful for assertions because it will display more details
// about the error than a simple `Nil` check.
func NoError(t *testing.T, err error, message ...any) {
	t.Helper()
	if !isNil(err) {
		fatalf(t, "unexpected error %+v", []any{err}, message)
	}
}

// Equal is a test helper to verify that two arguments are equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func Equal(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if !areEqual(expected, actual) {
		fatalf(t, "expected %v to equal %v", []any{actual, expected}, message)
	}
}

// NotEqual is a test helper to verify that two arguments are not equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func NotEqual(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if areEqual(expected, actual) {
		fatalf(t, "expected %v not to equal %v", []any{actual, expected}, message)
	}
}

// Nil is an assertion helper.
//
// It will test that the given value is nil, printing
// the value if the value has a string form.
func Nil(t *testing.T, v any, message ...any) {
	t.Helper()
	if !isNil(v) {
		fatalf(t, "expected value to be <nil>, was %v", []any{v}, message)
	}
}

// NotNil is an assertion helper.
//
// It will test that the given value is not nil.
func NotNil(t *testing.T, v any, message ...any) {
	t.Helper()
	if isNil(v) {
		fatalf(t, "expected value to not be <nil>", nil, message)
	}
}

// Matches is a test helper to verify that a string matches a given regular expression.
func Matches(t *testing.T, expression string, actual any, message ...any) {
	t.Helper()
	if !matches(expression, actual) {
		fatalf(t, "expected %v to match expression %v", []any{actual, expression}, message)
	}
}

// Empty asserts a slice is empty.
func Empty[A ~[]E, E any](t *testing.T, values A, message ...any) {
	t.Helper()
	if len(values) > 0 {
		fatalf(t, "expected %v to be empty", []any{values}, message)
	}
}

// NotEmpty asserts a slice is not empty.
func NotEmpty[A ~[]E, E any](t *testing.T, values A, message ...any) {
	t.Helper()
	if len(values) == 0 {
		fatalf(t, "expected slice not to be empty", []any{values}, message)
	}
}

// isNil returns if a given reference is nil, but also returning true
// if the reference is a valid typed pointer to nil, which may not strictly
// be equal to nil.
func isNil(object any) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

func areEqual(expected, actual any) bool {
	if isNil(expected) && isNil(actual) {
		return true
	}
	if (isNil(expected) && !isNil(actual)) || (!isNil(expected) && isNil(actual)) {
		return false
	}
	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}
	return reflect.DeepEqual(expected, actual)
}

func matches(expression string, actual any) bool {
	return regexp.MustCompile(expression).MatchString(fmt.Sprint(actual))
}

type blueDyeKey struct{}

// WithBlueDye adds a context value nonce to a context.
func WithBlueDye(ctx context.Context) context.Context {
	return context.WithValue(ctx, blueDyeKey{}, "test")
}

// HasBlueDye returns a boolean true if the context
// holds the blue dye variable.
func HasBlueDye(ctx context.Context) bool {
	return ctx.Value(blueDyeKey{}) != nil
}

// BlueDye asserts the context has the blue dye nonce on it.
//
// NOTE: we have to use this whack argument order because
// of linting rules around where "context.Context" has to appear.
func BlueDye(ctx context.Context, t *testing.T) {
	t.Helper()
	NotNil(t, ctx.Value(blueDyeKey{}))
}

// Any asserts a predicate matches any element in the values list.
func Any[A any](t *testing.T, values []A, pred func(A) bool, message ...any) {
	t.Helper()
	for _, v := range values {
		if pred(v) {
			return
		}
	}
	fatalf(t, "expected any value to match predicate", nil, message)
}

// All asserts a predicate matches all elements in the values list.
func All[A any](t *testing.T, values []A, pred func(A) bool, message ...any) {
	t.Helper()
	for _, v := range values {
		if !pred(v) {
			fatalf(t, "expected value to match predicate: %v", []any{v}, message)
		}
	}
}

// None asserts a predicate matches no elements in the values list.
func None[A any](t *testing.T, values []A, pred func(A) bool, message ...any) {
	t.Helper()
	for _, v := range values {
		if pred(v) {
			fatalf(t, "expected zero values to match predicate, value that matched: %v", []any{v}, message)
		}
	}
}

// Fail fails a test immediately.
func Fail(t *testing.T, message ...any) {
	t.Helper()
	t.Fatal(message...)
}

func fatalf(t *testing.T, format string, args, message []any) {
	t.Helper()
	if len(message) > 0 {
		t.Fatal(fmt.Sprintf(format, args...) + ": " + fmt.Sprint(message...))
	} else {
		t.Fatalf(format, args...)
	}
}
