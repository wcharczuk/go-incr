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

func itsDeepEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v to equal %v", actual, expected)
	}
}
