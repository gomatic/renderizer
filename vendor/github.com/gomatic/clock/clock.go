package clock

import (
	"time"
)

//
type TimeFunction func() time.Time

//
type Clock string

const (
	Default    = Clock("")
	Epoch      = Clock("1970-01-01 00:00:00.0 -0000 UTC")
	Format     = Clock("2006-01-02 15:04:05.987654321 -0000 UTC")
	Playground = Clock("2009-11-10 11:00:00.0 -0000 UTC")
	NineEleven = Clock("2001-09-11 12:46:40.0 -0000 UTC")
)

//
func Now(c Clock) TimeFunction {
	return c.MustTime()
}

//
func (c Clock) MustTime() TimeFunction {
	f, err := c.Time()
	if err != nil {
		panic(err)
	}
	return f
}

//
func (c Clock) Time() (TimeFunction, error) {
	if c == "" {
		return time.Now, nil
	}
	t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", string(c))
	if err != nil {
		return time.Now, err
	}
	return func() time.Time {
		return t
	}, nil
}

//
func (c Clock) UTC() TimeFunction {
	return c.MustTime()().UTC
}
