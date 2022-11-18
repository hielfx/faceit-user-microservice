package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// DateCheck - helper struct to check dates agains another
type DateCheck struct {
	Option DateCheckOption
	Value  time.Time
}

// DateCheckOption - date check options
type DateCheckOption int

const (
	DateCheckOptionBefore DateCheckOption = iota
	DateCheckOptionEquals
	DateCheckOptionAfter
)

func (d DateCheckOption) String() string {
	return [...]string{"Before", "Equals", "After"}[d]
}

func assertDateCheck(t *testing.T, dc DateCheck, name string, date time.Time) {
	var fn func(time.Time) bool
	switch dc.Option {
	case DateCheckOptionBefore:
		fn = date.Before
	case DateCheckOptionEquals:
		fn = date.Equal
	case DateCheckOptionAfter:
		fn = date.After
	default:
		return
	}

	assert.Truef(t, fn(dc.Value), "Expected %s to be after %s, but was %s", name, dc.Value, date)
}
