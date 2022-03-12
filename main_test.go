package incr

import (
	"reflect"
	"testing"
)

func itsEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if !areEqual(expected, actual) {
		t.Fatalf("expected %v, actual: %v", actual, expected)
	}
}

func itsNotEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if !areEqual(expected, actual) {
		t.Fatalf("expected: %v, actual: %v", expected, actual)
	}
}

func itsNil(t *testing.T, expected any) {
	t.Helper()
	if !isNil(expected) {
		t.Fatalf("expected %v to be nil", expected)
	}
}

func itsNotNil(t *testing.T, expected any) {
	t.Helper()
	if isNil(expected) {
		t.Fatalf("expected not to be nil")
	}
}

func itsAny[A comparable](t *testing.T, values []A, value A) {
	for _, v := range values {
		if v == value {
			return
		}
	}
	t.Fatalf("expected %v to be present in %v", value, values)
}

func itsEmpty[A any](t *testing.T, values []A) {
	t.Helper()
	if len(values) > 0 {
		t.Fatalf("expected %v to be empty", values)
	}
}

func itsNotEmpty[A any](t *testing.T, values []A) {
	t.Helper()
	if len(values) == 0 {
		t.Fatalf("expected %v to not be empty", values)
	}
}

func areEqual(expected, actual any) bool {
	if expected == nil && actual == nil {
		return true
	}
	if (expected == nil && actual != nil) || (expected != nil && actual == nil) {
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

func isNil(object interface{}) bool {
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
