package main

import (
	"context"
	"fmt"
	"time"

	incr "github.com/wcharczuk/go-incremental"
)

type Entry struct {
	Entry string
	Time  time.Time
}

func main() {
	now := time.Now()
	data := []Entry{
		{"0", now},
		{"1", now.Add(time.Second)},
		{"2", now.Add(2 * time.Second)},
		{"3", now.Add(3 * time.Second)},
		{"4", now.Add(4 * time.Second)},
	}

	i := incr.Var(data)
	output := incr.Map[[]Entry](
		i,
		func(entries []Entry) (output []string) {
			for _, e := range entries {
				if e.Time.Sub(now) > 2*time.Second {
					output = append(output, e.Entry)
				}
			}
			return
		},
	)

	_ = incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	)
	fmt.Println(output.Value())

	data = append(data, Entry{
		"5", now.Add(5 * time.Second),
	})
	_ = incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	)
	fmt.Println(output.Value())

	i.Set(data)
	_ = incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	)
	fmt.Println(output.Value())
}
