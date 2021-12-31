package incr

import (
	"reflect"
	"testing"
)

func itsEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v to equal %v", actual, expected)
	}
}

func itsNotEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if actual == expected {
		t.Fatalf("expected %v not to equal %v", actual, expected)
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

func itsDeepEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v to equal %v", actual, expected)
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
