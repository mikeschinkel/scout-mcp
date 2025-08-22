package fsfix

import (
	"testing"
)

type Fixture interface {
	Dir() string
	SetupWithParent(*testing.T, Fixture)
}
