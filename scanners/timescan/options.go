package timescan

import (
	"time"
)

type Options struct {
	TickMs  int
	Message string
}

func (o *Options) Tick() time.Duration {
	return time.Duration(o.TickMs) * time.Millisecond
}
