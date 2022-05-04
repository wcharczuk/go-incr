package incr

import (
	"fmt"
	"reflect"
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
