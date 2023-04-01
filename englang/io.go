package englang

import (
	"bytes"
	"fmt"
	"strings"
)

func ScanfContains(in string, format string, a ...*string) error {
	return scanfInner(in, format, a)
}

func ScanfPrefix(format string) string {
	items := strings.Split(format, "%s")
	if len(items) >= 1 {
		return items[0]
	}
	return ""
}

func ScanfSuffix(format string) string {
	items := strings.Split(format, "%s")
	if len(items) >= 1 {
		return items[len(items)-1]
	}
	return ""
}

func Scanf(in string, format string, an ...*string) error {
	begin := ""
	end := ""
	ab := make([]*string, len(an)+2)
	ab[0] = &begin
	ab[len(ab)-1] = &end
	copy(ab[1:1+len(an)], an)

	return scanfInner(in, format, ab)
}

func scanfInner(in string, format string, an []*string) error {
	items := strings.Split(format, "%s")
	expected := len(items)
	ai := 0
	for len(items) > 0 {
		index := strings.Index(in, items[0])
		if index == -1 {
			break
		}
		if index == 0 {
			*an[ai] = ""
			ai++
		} else {
			*an[ai] = in[0:index]
			ai++
		}
		in = in[index+len(items[0]):]
		items = items[1:]
	}
	*an[ai] = in
	if ai < expected {
		return fmt.Errorf("parsing error")
	}
	return nil
}

func ScanfStream(in []byte, i int, format string, an ...*string) (int, error) {
	begin := ""
	end := ""
	ab := make([]*string, len(an)+2)
	ab[0] = &begin
	ab[len(ab)-1] = &end
	copy(ab[1:1+len(an)], an)

	return scanfStreamInner(in, i, format, ab)
}

func scanfStreamInner(in []byte, i int, format string, an []*string) (int, error) {
	items := strings.Split(format, "%s")
	expected := len(items)
	marker := []byte(items[0])
	closure := []byte(items[expected-1])

	f := bytes.Index(in[i:], marker)
	if f == -1 {
		return -1, fmt.Errorf("not found")
	}
	e := bytes.Index(in[i+f:], closure)
	if e == -1 {
		return -1, fmt.Errorf("not found")
	}
	return i + f + e, scanfInner(string(in[i+f:i+f+e]), format, an)
}

func Printf(format string, an ...string) string {
	b := make([]string, len(an)+2)
	copy(b[1:], an)
	return sprintf(format, b)
}

func PrintfContains(format string, an ...string) string {
	return sprintf(format, an)
}

func sprintf(format string, an []string) string {
	items := strings.Split(format, "%s")
	ret := &strings.Builder{}
	ret.WriteString(an[0])
	for i := range items {
		ret.WriteString(items[i])
		ret.WriteString(an[i+1])
	}
	ret.WriteString(an[len(an)-1])
	return ret.String()
}
