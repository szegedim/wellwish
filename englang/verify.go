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

func IsAddress(address *string) bool {
	// We are liberal here, we enforce the country or territory.
	// The important is that the location of any Supreme Court is known.

	if IsBeingEdited(*address) {
		return true
	}
	territories := []string{
		"Afghanistan",
		"Albania",
		"Algeria",
		"Andorra",
		"Angola",
		"Antigua and Barbuda",
		"Argentina",
		"Armenia",
		"Australia",
		"Austria",
		"Azerbaijan",
		"Bahamas",
		"Bahrain",
		"Bangladesh",
		"Barbados",
		"Belarus",
		"Belgium",
		"Belize",
		"Benin",
		"Bhutan",
		"Bolivia",
		"Bosnia and Herzegovina",
		"Botswana",
		"Brazil",
		"Brunei",
		"Bulgaria",
		"Burkina Faso",
		"Burundi",
		"Cabo Verde",
		"Cambodia",
		"Cameroon",
		"Canada",
		"Central African Republic",
		"Chad",
		"Chile",
		"China",
		"Colombia",
		"Comoros",
		"Congo, Democratic Republic of the",
		"Congo, Republic of the",
		"Costa Rica",
		"Cote d'Ivoire",
		"Croatia",
		"Cuba",
		"Cyprus",
		"Czech Republic",
		"Denmark",
		"Djibouti",
		"Dominica",
		"Dominican Republic",
		"East Timor (Timor-Leste)",
		"Ecuador",
		"Egypt",
		"El Salvador",
		"Equatorial Guinea",
		"Eritrea",
		"Estonia",
		"Eswatini",
		"Ethiopia",
		"Fiji",
		"Finland",
		"France",
		"Gabon",
		"Gambia",
		"Georgia",
		"Germany",
		"Ghana",
		"Greece",
		"Grenada",
		"Guatemala",
		"Guinea",
		"Guinea-Bissau",
		"Guyana",
		"Haiti",
		"Honduras",
		"Hungary",
		"Iceland",
		"India",
		"Indonesia",
		"Iran",
		"Iraq",
		"Ireland",
		"Israel",
		"Italy",
		"Jamaica",
		"Japan",
		"Jordan",
		"Kazakhstan",
		"Kenya",
		"Kiribati",
		"Korea, North",
		"Korea, South",
		"Kosovo",
		"Kuwait",
		"Kyrgyzstan",
		"Laos",
		"Latvia",
		"Lebanon",
		"Lesotho",
		"Liberia",
		"Libya",
		"Liechtenstein",
		"Lithuania",
		"Luxembourg",
		"Madagascar",
		"Malawi",
		"Malaysia",
		"Maldives",
		"Mali",
		"Malta",
		"Marshall Islands",
		"Mauritania",
		"Mauritius",
		"Mexico",
		"Micronesia",
		"Moldova",
		"Monaco",
		"Mongolia",
		"Montenegro",
		"Morocco",
		"Mozambique",
		"Myanmar",
		"Burma",
		"Namibia",
		"Nauru",
		"Nepal",
		"Netherlands",
		"New Zealand",
		"Nicaragua",
		"Niger",
		"Nigeria",
		"North Macedonia",
		"Norway",
		"Oman",
		"Pakistan",
		"Palau",
		"Panama",
		"Papua New Guinea",
		"Paraguay",
		"Peru",
		"Philippines",
		"Poland",
		"Portugal",
		"Qatar",
		"Romania",
		"Russia",
		"Rwanda",
		"Saint Kitts and Nevis",
		"Saint Lucia",
		"Saint Vincent and the Grenadines",
		"Samoa",
		"San Marino",
		"Sao Tome and Principe",
		"Saudi Arabia",
		"Senegal",
		"Serbia",
		"Seychelles",
		"Sierra Leone",
		"Singapore",
		"Slovakia",
		"Slovenia",
		"Solomon Islands",
		"Somalia",
		"South Africa",
		"South Sudan",
		"Spain",
		"Sri Lanka",
		"Sudan",
		"Suriname",
		"Sweden",
		"Switzerland",
		"Syria",
		"Taiwan",
		"Tajikistan",
		"Tanzania",
		"Thailand",
		"Togo",
		"Tonga",
		"Trinidad and Tobago",
		"Tunisia",
		"Turkey",
		"Turkmenistan",
		"Tuvalu",
		"Uganda",
		"Ukraine",
		"United Arab Emirates",
		"United Kingdom",
		"United States",
		"Uruguay",
		"Uzbekistan",
		"Vanuatu",
		"Vatican City",
		"Venezuela",
		"Vietnam",
		"Yemen",
		"Zambia",
		"Zimbabwe"}
	check := *address
	check = strings.ReplaceAll(check, "\v", "")
	check = strings.ReplaceAll(check, "\v", "")
	for _, x := range territories {
		if strings.HasSuffix(check, x) {
			return true
		}
	}
	candidate := ""
	for _, x := range territories {
		for i := 1; i < len(x); i++ {
			if strings.HasPrefix(x, (check)[len(check)-i:]) {
				if candidate != "" {
					return true
				}
				candidate = x
			}
		}
	}
	return candidate != ""
	//*address = fmt.Sprintf("%s"+format+"%s", addressLines, city, state, zip, country)
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
			dot := false
			for _, c := range factor {
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
					unit = unit + string(c)
				}
				if c >= '0' && c <= '9' {
					value = value*10 + uint64(c-'0')
					if dot {
						residual = residual + 1
					}
				}
				if c == ',' {
					continue
				}
				if c == '.' {
					dot = true
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
			s = s[:len(s)-evaluatedResidual] + "." + s[len(s)-evaluatedResidual:]
		}
	}
	return s
}
