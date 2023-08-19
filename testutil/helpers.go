package testutil

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

// Fatalf is a helper to fatal a test with optional message components.
func Fatalf(t *testing.T, format string, args, message []any) {
	t.Helper()
	if len(message) > 0 {
		t.Fatal(fmt.Sprintf(format, args...) + ": " + fmt.Sprint(message...))
	} else {
		t.Fatalf(format, args...)
	}
}

// ItsEqual is a test helper to verify that two arguments are equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func ItsEqual(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if !areEqual(expected, actual) {
		Fatalf(t, "expected %v to equal %v", []any{actual, expected}, message)
	}
}

// ItsNotEqual is a test helper to verify that two arguments are not equal.
//
// You can use it to build up other assertions, such as for length or not-nil.
func ItsNotEqual(t *testing.T, expected, actual any, message ...any) {
	t.Helper()
	if areEqual(expected, actual) {
		Fatalf(t, "expected %v not to equal %v", []any{actual, expected}, message)
	}
}

// ItsNil is an assertion helper.
//
// It will test that the given value is nil, printing
// the value if the value has a string form.
func ItsNil(t *testing.T, v any, message ...any) {
	t.Helper()
	if !Nil(v) {
		Fatalf(t, "expected value to be <nil>, was %v", []any{v}, message)
	}
}

// ItsNotNil is an assertion helper.
//
// It will test that the given value is not nil.
func ItsNotNil(t *testing.T, v any, message ...any) {
	t.Helper()
	if Nil(v) {
		Fatalf(t, "expected value to not be <nil>", nil, message)
	}
}

// ItMatches is a test helper to verify that a string matches a given regular expression.
func ItMatches(t *testing.T, expression string, actual any, message ...any) {
	t.Helper()
	if !matches(expression, actual) {
		Fatalf(t, "expected %v to match expression %v", []any{actual, expression}, message)
	}
}

// Nil returns if a given reference is nil, but also returning true
// if the reference is a valid typed pointer to nil, which may not strictly
// be equal to nil.
func Nil(object any) bool {
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
	if Nil(expected) && Nil(actual) {
		return true
	}
	if (Nil(expected) && !Nil(actual)) || (!Nil(expected) && Nil(actual)) {
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

// ItsBlueDye asserts the context has the blue dye nonce on it.
//
// NOTE: we have to use this whack order because of linting rules
// around where "context.Context" has to appear.
func ItsBlueDye(ctx context.Context, t *testing.T) {
	t.Helper()
	ItsNotNil(t, ctx.Value(blueDyeKey{}))
}
