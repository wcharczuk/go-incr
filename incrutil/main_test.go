package incrutil

import (
	"context"
	"os"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func testContext() context.Context {
	ctx := context.Background()
	if os.Getenv("DEBUG") != "" {
		ctx = incr.WithTracing(ctx)
	}
	ctx = testutil.WithBlueDye(ctx)
	return ctx
}
