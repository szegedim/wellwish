package englang

import (
	"fmt"
	"strings"
)

func IsBeingEdited(s string) bool {
	if s == "" {
		return true
	}
	if strings.Contains(s, "�") {
		return true
	}
	return false
}

func IsEmail(at string) bool {
	if IsBeingEdited(at) {
		return true
	}
	if !strings.Contains(at, "@") {
		return false
	}
	if !strings.Contains(at, ".") {
		return false
	}
	for _, c := range at {
		found := false
		if c >= 'a' && c <= 'z' {
			found = true
		}
		if c >= '0' && c <= '9' {
			found = true
		}
		if c == '@' {
			found = true
		}
		if c == '.' {
			found = true
		}
		if c == '_' {
			found = true
		}
		if c == '\v' {
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}

func IsCompany(company string) bool {
	if IsBeingEdited(company) {
		return true
	}
	company = strings.TrimSpace(company)
	for _, c := range company {
		found := false
		if c >= 'a' && c <= 'z' {
			found = true
		}
		if c >= 'A' && c <= 'Z' {
			found = true
		}
		if c == '&' {
			found = true
		}
		if c == '.' {
			found = true
		}
		if c == ' ' {
			found = true
		}
		if c == '\v' {
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}

func IsNumber(number string) bool {
	if IsBeingEdited(number) {
		return true
	}
	for _, c := range number {
		found := false
		if c == '.' {
			found = true
		}
		if c == ',' {
			found = true
		}
		if c >= '0' && c <= '9' {
			found = true
		}
		if c == '\v' {
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}

func IsZip(zip string) bool {
	if IsBeingEdited(zip) {
		return true
	}
	zip = strings.TrimSpace(zip)
	for _, c := range zip {
		found := false
		if c == '.' {
			found = true
		}
		if c == ',' {
			found = true
		}
		if c == '-' {
			found = true
		}
		if c >= '0' && c <= '9' {
			found = true
		}
		if c == '\v' {
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}

func IsCountry(country string) bool {
	if IsBeingEdited(country) {
		return true
	}
	country = strings.TrimSpace(country)
	var countries = []string{"USA", "United States", "Canada", "Mexico"}
	for _, i := range countries {
		if strings.HasPrefix(i, country) {
			return true
		}
	}
	return false
}

func IsState(state string) bool {
	_, ret := FixState(state)
	return ret
}

func FixState(state string) (string, bool) {
	if IsBeingEdited(state) {
		return state, true
	}
	var states = []string{"California", "CA"}
	for _, i := range states {
		if state == i {
			return i, true
		}
	}
	return "", false
}

func IsAddress(address *string) bool {
	addressLines := ""
	city := ""
	state := ""
	zip := ""
	country := ""
	const format = ", %s, %s, %s,"
	err := ScanfContains(*address, format, &addressLines, &city, &state, &zip, &country)
	if IsBeingEdited(addressLines) {
		return true
	}
	if IsBeingEdited(city) {
		return true
	}
	if IsBeingEdited(state) {
		return true
	}
	if !IsState(state) {
		return false
	}
	if IsBeingEdited(zip) {
		return true
	}
	if !IsZip(zip) {
		return false
	}
	if IsBeingEdited(country) {
		return true
	}
	if !IsCountry(country) {
		return false
	}
	if err != nil {
		return false
	}
	*address = fmt.Sprintf("%s"+format+"%s", addressLines, city, state, zip, country)
	return err == nil
}

func Evaluate(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	factors := strings.Split(s, "multipliedby")
	if len(factors) > 1 {
		evaluatedUnit := ""
		evaluatedValue := uint64(0)
		evaluatedResidual := int(0)
		for _, factor := range factors {
			unit := ""
			value := uint64(0)
			residual := int(0)
			for _, c := range factor {
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
					unit = unit + string(c)
				}
				if c >= '0' && c <= '9' {
					value = value*10 + uint64(c-'0')
					if residual > 0 {
						residual = residual + 1
					}
				}
				if c == ',' {
					continue
				}
				if c == '.' {
					residual = 1
				}
			}
			evaluatedUnit = evaluatedUnit + unit
			if evaluatedValue == 0 {
				evaluatedValue = value
				evaluatedResidual = residual
			} else {
				evaluatedValue = evaluatedValue * value
				evaluatedResidual = evaluatedResidual + residual
			}
		}
		s = fmt.Sprintf("%s %d", evaluatedUnit, evaluatedValue)
		if evaluatedResidual > 0 {
			s = s[:len(s)+1-evaluatedResidual] + "." + s[len(s)+1-evaluatedResidual:]
		}
	}
	return s
}
