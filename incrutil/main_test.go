package incrutil

import (
	"context"

	"github.com/wcharczuk/go-incr/testutil"
)

func testContext() context.Context {
	ctx := context.Background()
	ctx = testutil.WithBlueDye(ctx)
	return ctx
}
