package generator

import (
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type DateGenerator struct {
	Generator
}

func (g DateGenerator) Date(args ...string) any {
	if len(args) > 0 {
		format := args[0]
		switch format {
		case "ANSIC":
			return gofakeit.Date().Format(time.ANSIC)
		case "UnixDate":
			return gofakeit.Date().Format(time.UnixDate)
		case "RubyDate":
			return gofakeit.Date().Format(time.RubyDate)
		case "RFC822":
			return gofakeit.Date().Format(time.RFC822)
		case "RFC822Z":
			return gofakeit.Date().Format(time.RFC822Z)
		case "RFC850":
			return gofakeit.Date().Format(time.RFC850)
		case "RFC1123":
			return gofakeit.Date().Format(time.RFC1123)
		case "RFC1123Z":
			return gofakeit.Date().Format(time.RFC1123Z)
		case "RFC3339":
			return gofakeit.Date().Format(time.RFC3339)
		case "RFC3339Nano":
			return gofakeit.Date().Format(time.RFC3339Nano)
		default:
			return gofakeit.Date().Format(formatToGoFormat(format))
		}
	}

	return gofakeit.Date()
}

func (g DateGenerator) Timestamp() int64 {
	return gofakeit.DateRange(time.Unix(0, 0), gofakeit.FutureDate()).Unix()
}

func (g DateGenerator) Day() int {
	return gofakeit.Day()
}

func (g DateGenerator) Month(args ...string) any {
	if len(args) > 0 {
		if args[0] == "string" {
			return gofakeit.MonthString()
		}
	}

	return gofakeit.Month()
}

func (g DateGenerator) Year() int {
	return gofakeit.Year()
}

func (g DateGenerator) WeekDay() string {
	return gofakeit.WeekDay()
}

func (g DateGenerator) Future() time.Time {
	return gofakeit.FutureDate()
}

func (g DateGenerator) Past() time.Time {
	return gofakeit.PastDate()
}

func formatToGoFormat(format string) string {
	format = strings.Replace(format, "ddd", "_2", -1)
	format = strings.Replace(format, "dd", "02", -1)
	format = strings.Replace(format, "d", "2", -1)

	format = strings.Replace(format, "HH", "15", -1)

	format = strings.Replace(format, "hh", "03", -1)
	format = strings.Replace(format, "h", "3", -1)

	format = strings.Replace(format, "mm", "04", -1)
	format = strings.Replace(format, "m", "4", -1)

	format = strings.Replace(format, "ss", "05", -1)
	format = strings.Replace(format, "s", "5", -1)

	format = strings.Replace(format, "yyyy", "2006", -1)
	format = strings.Replace(format, "yy", "06", -1)
	format = strings.Replace(format, "y", "06", -1)

	format = strings.Replace(format, "SSS", "000", -1)

	format = strings.Replace(format, "a", "pm", -1)
	format = strings.Replace(format, "aa", "PM", -1)

	format = strings.Replace(format, "MMMM", "January", -1)
	format = strings.Replace(format, "MMM", "Jan", -1)
	format = strings.Replace(format, "MM", "01", -1)
	format = strings.Replace(format, "M", "1", -1)

	format = strings.Replace(format, "ZZ", "-0700", -1)

	if !strings.Contains(format, "Z07") {
		format = strings.Replace(format, "Z", "-07", -1)
	}

	format = strings.Replace(format, "zz:zz", "Z07:00", -1)
	format = strings.Replace(format, "zzzz", "Z0700", -1)
	format = strings.Replace(format, "z", "MST", -1)

	format = strings.Replace(format, "EEEE", "Monday", -1)
	format = strings.Replace(format, "E", "Mon", -1)

	return format
}
