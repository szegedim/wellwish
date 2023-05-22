package englang

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Synonym(s string, t string) bool {
	// Returns, if the two Englang statements are equivalent
	// Example: Context is voltage. It is 3.3V == The voltage is 3.3V
	return s == t
}

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

func Scanf1(in string, format string, an ...*string) error {
	begin := ""
	end := ""
	ab := make([]*string, len(an)+2)
	ab[0] = &begin
	ab[len(ab)-1] = &end
	copy(ab[1:1+len(an)], an)

	ret := scanfInner(in, format, ab)
	if begin != "" || end != "" {
		return fmt.Errorf("no match")
	}
	return ret
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

func Println(format string, an ...string) string {
	b := make([]string, len(an)+2)
	copy(b[1:], an)
	return sprintf(format+"\n", b)
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

func Decimal(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		i = 0
	}
	return i
}

func DecimalString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ReadWith(in io.Reader, closure string) string {
	var ret = bytes.Buffer{}
	var p = make([]byte, 1)
	for {
		n, _ := in.Read(p)
		if n == 0 {
			return ""
		}
		ret.WriteString(string(p))
		if strings.HasSuffix(ret.String(), closure) {
			return ret.String()
		}
	}
}
