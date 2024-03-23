package generator

import (
	"math"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
)

type NumberGenerator struct {
	Generator
}

func randRangeInt(min int, max int) int {
	return gofakeit.IntRange(min, max)
}
func randRangeFloat(min float64, max float64) float64 {
	return gofakeit.Float64Range(min, max)
}

func (g NumberGenerator) Number() int {
	return gofakeit.Int()
}

func (g NumberGenerator) Float() float64 {
	return gofakeit.Float64()
}

func (g NumberGenerator) FloatRange(min string, max string) float64 {
	if min == "x" && max == "x" {
		return g.Float()
	}

	if min == "x" {
		fMax, _ := strconv.ParseFloat(max, 64)
		return randRangeFloat(0, fMax)
	} else if max == "x" {
		fMin, _ := strconv.ParseFloat(min, 64)
		return randRangeFloat(fMin, math.MaxInt)
	} else {
		fMin, _ := strconv.ParseFloat(min, 64)
		fMax, _ := strconv.ParseFloat(max, 64)
		return randRangeFloat(fMin, fMax)
	}
}

func (g NumberGenerator) Decimal(p string, min string, max string) string {
	prec, _ := strconv.Atoi(p)
	if (min == "" && max == "") || (min == "x" && max == "x") {
		return strconv.FormatFloat(g.Float(), 'f', prec, 32)
	}

	if min == "x" {
		fMax, _ := strconv.ParseFloat(max, 32)
		return strconv.FormatFloat(g.FloatRange("x", strconv.FormatFloat(fMax, 'f', -1, 32)), 'f', prec, 32)
	} else if max == "x" {
		fMin, _ := strconv.ParseFloat(min, 32)
		return strconv.FormatFloat(g.FloatRange(strconv.FormatFloat(fMin, 'f', -1, 32), "x"), 'f', prec, 32)
	} else {
		fMin, _ := strconv.ParseFloat(min, 32)
		fMax, _ := strconv.ParseFloat(max, 32)
		return strconv.FormatFloat(g.FloatRange(strconv.FormatFloat(fMin, 'f', -1, 32), strconv.FormatFloat(fMax, 'f', -1, 32)), 'f', prec, 32)
	}
}

func (g NumberGenerator) Int(min string, max string) int {
	if min == "x" && max == "x" {
		return g.Number()
	}

	if min == "x" {
		iMax, _ := strconv.Atoi(max)
		return randRangeInt(0, iMax)
	} else if max == "x" {
		iMin, _ := strconv.Atoi(min)
		return randRangeInt(iMin, math.MaxInt)
	} else {
		iMin, _ := strconv.Atoi(min)
		iMax, _ := strconv.Atoi(max)
		return randRangeInt(iMin, iMax)
	}
}
