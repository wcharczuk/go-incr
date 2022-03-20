package incr

import (
	"context"
	"testing"
	"time"
)

func Test_Delay(t *testing.T) {
	now := time.Now()
	nowProvider := func() time.Time {
		return now
	}

	i := Var(1.0)
	_ = i.Stabilize(context.Background())
	d := Delay(
		Map[float64](
			i,
			func(i float64) float64 {
				return i * 2.0
			},
		),
		100*time.Millisecond,
	)
	d.(*delayIncr[float64]).n.now = nowProvider
	err := d.Stabilize(context.Background())
	itsNil(t, err)
	itsEqual(t, 2.0, d.Value())

	i.Set(2.0)
	_ = i.Stabilize(context.Background())

	err = d.Stabilize(context.Background())
	itsNil(t, err)
	itsEqual(t, 2.0, d.Value())

	now = now.Add(200 * time.Millisecond)

	err = d.Stabilize(context.Background())
	itsNil(t, err)
	itsEqual(t, 4.0, d.Value())
}
