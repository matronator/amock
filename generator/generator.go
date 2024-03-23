package generator

import "github.com/brianvoe/gofakeit/v7"

type Generator struct{}

type BoolGenerator struct {
	Generator
}

func (g BoolGenerator) Bool() bool {
	return gofakeit.Bool()
}
