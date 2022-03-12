package incr

import (
	"math"
	"reflect"
	"testing"
)

//
// helpers
//

// epsilon returns a function that returns true
// if the absolute difference of two values is greater
// than a given delta.
func epsilon[A int | float32 | float64](delta A) func(A, A) bool {
	return func(v0, v1 A) bool {
		return math.Abs(float64(v1)-float64(v0)) > float64(delta)
	}
}

// addConst returns a map fn that adds a constant value
// to a given input
func addConst[A int | float32 | float64](v A) func(A) A {
	return func(v0 A) A {
		return v0 + v
	}
}

// add is a map2 fn that adds two values and returns the result
func add[A int | float32 | float64](v0, v1 A) A {
	return v0 + v1
}

func makeFib(height int) (output Incr[int]) {
	prev2 := Return(0) // 0
	prev := Return(1)  // 1
	current := Map2(   // 2
		prev2,
		prev,
		add[int],
	)
	for x := 3; x < height; x++ {
		prev2 = prev
		prev = current
		current = Map2(
			prev2,
			prev,
			add[int],
		)
	}
	output = Map2(
		prev,
		current,
		add[int],
	)
	return
}

//
// assertions
//

func itsEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if !areEqual(expected, actual) {
		t.Fatalf("equal; expected %v, actual: %v", actual, expected)
	}
}

func itsNotEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if areEqual(expected, actual) {
		t.Fatalf("not equal; expected: %v, actual: %v", expected, actual)
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
