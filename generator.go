package main

import (
	"reflect"
	"strconv"
	"strings"

	"amock/generator"
	"github.com/oriser/regroup"
)

var FieldPattern = regroup.MustCompile(`(?P<type>[a-z]+)(?P<subtype>\.[a-z]+)?(?P<params>:.*)?`)
var NumberRangePattern = regroup.MustCompile(`(?P<min>(-?[0-9]+(\.[0-9]+)?)|x)?-(?P<max>(-?[0-9]+(\.[0-9]+)?)|x)?`)

type Field struct {
	Type     string `regroup:"type" json:"type"`
	Subtype  string `regroup:"subtype" json:"subtype"`
	Params   string `regroup:"params" json:"params"`
	Required bool   `json:"required"`
	Nullable bool   `json:"nullable"`
}

type FieldOptions struct {
	Required bool
	Nullable bool
	Children bool
}

func GenerateField(fieldName string, field string, table *Table, options FieldOptions) (any, *Table) {
	f := *GetFieldType(field)
	f.Required = options.Required
	f.Nullable = options.Nullable
	table.Definition[fieldName] = &f

	return GenerateEntityField(f, table)
}

func GenerateEntityField(field Field, table *Table) (any, *Table) {
	gen := GetGenerator(field.Type, field.Subtype)
	var paramStr string
	var params []string

	if len(field.Params) > 1 {
		paramStr = strings.TrimLeft(field.Params, ":")

		if field.Type == "number" {
			params = strings.Split(paramStr, ",")
			for i, p := range params {
				if strings.Contains(p, "-") {
					groups, _ := NumberRangePattern.Groups(p)
					if groups["min"] == "" && groups["max"] == "" {
						params[i] = "x-x"
					} else {
						params[i] = groups["min"]
						params = append(params, groups["max"])
					}
				}
			}
		} else {
			params = strings.Split(paramStr, ",")
		}
	} else if field.Type == "id" && field.Subtype != "uuid" {
		params = []string{strconv.Itoa(int(table.LastAutoID))}
		table.LastAutoID = table.LastAutoID + 1
	}

	if len(params) > 0 {
		params2 := make([]reflect.Value, len(params))
		for i, p := range params {
			params2[i] = reflect.ValueOf(p)
		}
		return reflect.ValueOf(gen).Call(params2)[0].Interface(), table
	}

	return reflect.ValueOf(gen).Call([]reflect.Value{})[0].Interface(), table
}

func GetFieldType(field string) *Field {
	f := &Field{}

	err := FieldPattern.MatchToTarget(field, f)

	if err != nil {
		panic(err)
	}

	var subtype string

	if len(f.Subtype) > 1 {
		subtype = strings.TrimLeft(f.Subtype, ".")
	}

	f.Subtype = subtype

	return f
}

type GeneratorFunc any

type GeneratorMap map[string]GeneratorFunc

var Generators = map[string]GeneratorMap{
	"string": {
		"root":       generator.StringGenerator{}.String,
		"name":       generator.StringGenerator{}.FullName,
		"firstname":  generator.StringGenerator{}.FirstName,
		"lastname":   generator.StringGenerator{}.LastName,
		"email":      generator.StringGenerator{}.Email,
		"url":        generator.StringGenerator{}.Url,
		"ip":         generator.StringGenerator{}.Ip,
		"ipv6":       generator.StringGenerator{}.Ipv6,
		"username":   generator.StringGenerator{}.Username,
		"password":   generator.StringGenerator{}.Password,
		"phone":      generator.StringGenerator{}.Phone,
		"zip":        generator.StringGenerator{}.Zip,
		"country":    generator.StringGenerator{}.Country,
		"city":       generator.StringGenerator{}.City,
		"street":     generator.StringGenerator{}.Street,
		"streetName": generator.StringGenerator{}.StreetName,
		"state":      generator.StringGenerator{}.State,
		"company":    generator.StringGenerator{}.Company,
		"bitcoin":    generator.StringGenerator{}.Bitcoin,
		"color":      generator.StringGenerator{}.Color,
		"word":       generator.StringGenerator{}.Word,
		"sentence":   generator.StringGenerator{}.Sentence,
		"paragraph":  generator.StringGenerator{}.Paragraph,
	},
	"number": {
		"root":    generator.NumberGenerator{}.Number,
		"int":     generator.NumberGenerator{}.Int,
		"decimal": generator.NumberGenerator{}.Decimal,
		"float":   generator.NumberGenerator{}.Float,
		"range":   generator.NumberGenerator{}.FloatRange,
	},
	"date": {
		"root":      generator.DateGenerator{}.Date,
		"timestamp": generator.DateGenerator{}.Timestamp,
		"day":       generator.DateGenerator{}.Day,
		"month":     generator.DateGenerator{}.Month,
		"year":      generator.DateGenerator{}.Year,
		"weekday":   generator.DateGenerator{}.WeekDay,
		"future":    generator.DateGenerator{}.Future,
		"past":      generator.DateGenerator{}.Past,
	},
	"bool": {
		"root": generator.BoolGenerator{}.Bool,
	},
	"enum": {
		"root": generator.EnumGenerator{}.Enum,
	},
	"id": {
		"root":     generator.IDGenerator{}.Sequence,
		"sequence": generator.IDGenerator{}.Sequence,
		"uuid":     generator.IDGenerator{}.UUID,
	},
}

func GetGenerator(t string, subtype string) any {
	var gen any

	if subtype == "" {
		gen = Generators[t]["root"]
	} else {
		gen = Generators[t][subtype]
	}

	return gen
}
