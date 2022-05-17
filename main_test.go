package incr

import (
	"context"
	"fmt"
	"math"
	"os"
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

func testContext() context.Context {
	ctx := context.Background()
	if os.Getenv("DEBUG") != "" {
		ctx = WithTracing(ctx)
	}
	ctx = withBlueDye(ctx)
	return ctx
}

type blueDyeKey struct{}

func withBlueDye(ctx context.Context) context.Context {
	return context.WithValue(ctx, blueDyeKey{}, "test")
}

// itsBlueDye asserts the context has the blue dye trace on it.
//
// NOTE: we have to use this whack order because of linting rules.
func itsBlueDye(ctx context.Context, t *testing.T) {
	t.Helper()
	ItsNotNil(t, ctx.Value(blueDyeKey{}))
}

// epsilon returns a function that returns true
// if the absolute difference of two values is less
// than or equal to a given delta.
//
// this serves to implement a cutoff function, where we should
// cutoff the computation if a difference is sufficiently small.
func epsilon(delta float64) func(float64, float64) bool {
	return func(v0, v1 float64) bool {
		return math.Abs(v1-v0) <= delta
	}
}

// addConst returs a map fn that adds a constant value
// to a given input
func addConst(v float64) func(float64) float64 {
	return func(v0 float64) float64 {
		return v0 + v
	}
}

// add is a map2 fn that adds two values and returns the result
func add[T Ordered](v0, v1 T) T {
	return v0 + v1
}

type mockBareNode struct {
	n *Node
}

func (mn *mockBareNode) Node() *Node {
	if mn.n == nil {
		mn.n = NewNode()
	}
	return mn.n
}

func newHeightIncr(height int) *heightIncr {
	return &heightIncr{
		n: &Node{
			id:     NewIdentifier(),
			height: height,
		},
	}
}

func newHeightIncrLabel(height int, label string) *heightIncr {
	return &heightIncr{
		n: &Node{
			id:     NewIdentifier(),
			height: height,
			label:  label,
		},
	}
}

type heightIncr struct {
	Incr[struct{}]
	n *Node
}

func (hi heightIncr) Node() *Node {
	return hi.n
}
