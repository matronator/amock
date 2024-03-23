package generator

import (
	"fmt"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
)

type StringGenerator struct {
	Generator
}

func (g StringGenerator) String() string {
	return gofakeit.Regex("[A-z\\-_+?&*$@/!=#]{3,16}")
}

func (g StringGenerator) FullName() string {
	return gofakeit.Name()
}

func (g StringGenerator) FirstName() string {
	return gofakeit.FirstName()
}

func (g StringGenerator) LastName() string {
	return gofakeit.LastName()
}

func (g StringGenerator) Email() string {
	return gofakeit.Email()
}

func (g StringGenerator) Url() string {
	return gofakeit.URL()
}

func (g StringGenerator) Ip() string {
	return gofakeit.IPv4Address()
}

func (g StringGenerator) Ipv6() string {
	return gofakeit.IPv6Address()
}

func (g StringGenerator) Username() string {
	return gofakeit.Username()
}

func (g StringGenerator) Password() string {
	return gofakeit.Password(true, true, true, true, false, 16)
}

func (g StringGenerator) Phone() string {
	return gofakeit.Phone()
}

func (g StringGenerator) Zip() string {
	return gofakeit.Zip()
}

func (g StringGenerator) Country(args ...string) string {
	if len(args) > 0 {
		if args[0] == "short" {
			return gofakeit.CountryAbr()
		}
	}

	return gofakeit.Country()
}

func (g StringGenerator) City() string {
	return gofakeit.City()
}

func (g StringGenerator) Street() string {
	return gofakeit.Street()
}

func (g StringGenerator) StreetName() string {
	return gofakeit.StreetName()
}

func (g StringGenerator) State(args ...string) string {
	if len(args) > 0 {
		if args[0] == "short" {
			return gofakeit.StateAbr()
		}
	}

	return gofakeit.State()
}

func (g StringGenerator) Company() string {
	return gofakeit.Company()
}

func (g StringGenerator) Bitcoin() string {
	return gofakeit.BitcoinAddress()
}

func (g StringGenerator) Color(args ...string) string {
	if len(args) > 0 {
		if args[0] == "hex" {
			return gofakeit.HexColor()
		}

		if args[0] == "safe" {
			return gofakeit.SafeColor()
		}

		if args[0] == "rgb" {
			rgb := gofakeit.RGBColor()
			return fmt.Sprintf("rgb(%d, %d, %d)", rgb[0], rgb[1], rgb[2])
		}
	}

	return gofakeit.Color()
}

func (g StringGenerator) Word() string {
	return gofakeit.Word()
}

func (g StringGenerator) Sentence(args ...string) string {
	if len(args) > 0 {
		n, _ := strconv.Atoi(args[0])

		return gofakeit.Sentence(n)
	}

	return gofakeit.Sentence(3)
}

func (g StringGenerator) Paragraph(args ...string) string {
	if len(args) <= 0 {
		return gofakeit.Paragraph(3, 4, 12, "\n\n")
	}

	n, _ := strconv.Atoi(args[0])

	return gofakeit.Paragraph(n, 4, 12, "\n\n")
}
