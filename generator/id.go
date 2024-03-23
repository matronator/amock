package generator

import (
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
)

type IDGenerator struct {
	Generator
}

func (g IDGenerator) UUID() string {
	return gofakeit.UUID()
}

func (g IDGenerator) Sequence(args ...string) uint {
	if len(args) > 0 {
		id, _ := strconv.Atoi(args[0])

		return uint(id)
	}

	return 1
}
