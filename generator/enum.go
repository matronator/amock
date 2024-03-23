package generator

import "github.com/brianvoe/gofakeit/v7"

type EnumGenerator struct {
	Generator
}

func (g EnumGenerator) Enum(args ...string) string {
	return gofakeit.RandomString(args)
}
