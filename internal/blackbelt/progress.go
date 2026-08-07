package blackbelt

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

var activityFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type activity struct {
	label  string
	stream *os.File
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

func startActivity(label string) *activity {
	return newActivity(label, os.Stderr, terminal(os.Stderr))
}

func newActivity(label string, stream *os.File, interactive bool) *activity {
	a := &activity{label: label, stream: stream}
	if !interactive {
		return a
	}
	a.stop = make(chan struct{})
	a.done = make(chan struct{})
	a.draw(activityFrames[0], true)
	go func() {
		defer close(a.done)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		frame := 1
		for {
			select {
			case <-a.stop:
				return
			case <-ticker.C:
				a.draw(activityFrames[frame%len(activityFrames)], true)
				frame++
			}
		}
	}()
	return a
}

func (a *activity) finish(success bool) {
	if a.done == nil {
		return
	}
	a.once.Do(func() { close(a.stop) })
	<-a.done
	symbol := "✓"
	green := true
	if !success {
		symbol = "✗"
		green = false
	}
	a.draw(symbol, green)
	fmt.Fprintln(a.stream)
}

func (a *activity) draw(symbol string, green bool) {
	color, reset := "", ""
	if os.Getenv("NO_COLOR") == "" {
		color, reset = "\033[31m", "\033[0m"
		if green {
			color = "\033[32m"
		}
	}
	fmt.Fprintf(a.stream, "\r\033[K%s%s %s%s", color, symbol, a.label, reset)
}

func mustWithActivity(ctx context.Context, r runner, label, name string, args ...string) (string, error) {
	progress := startActivity(label)
	output, err := must(ctx, r, name, args...)
	progress.finish(err == nil)
	return output, err
}
