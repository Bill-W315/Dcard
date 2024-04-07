package api

import (
	"errors"
	"net/http"
	"strconv"
)

func validateAd(ad Ad) error {
	//Required fields
	//Title is empty string
	if ad.Title == "" {
		return errors.New("ad title cannot be empty")
	}

	//startAt is empty
	if ad.StartAt.IsZero() {
		return errors.New("startAt cannot be empty")
	}

	//endAt is empty
	if ad.EndAt.IsZero() {
		return errors.New("endAt cannot be empty")
	}

	//startAt > endAt
	if ad.StartAt.After(ad.EndAt) {
		return errors.New("startAt is after endAt")
	}

	//Optional fields
	//Missing one value
	if (ad.Conditions.AgeStart == 0 && ad.Conditions.AgeEnd != 0) || (ad.Conditions.AgeStart != 0 && ad.Conditions.AgeEnd == 0) {
		return errors.New("ageStart or ageEnd is missing")
	}

	//ageStart > ageEnd
	if ad.Conditions.AgeStart > ad.Conditions.AgeEnd {
		return errors.New("ageStart > ageEnd")
	}

	//ageStart < 1
	if ad.Conditions.AgeStart != 0 && ad.Conditions.AgeStart < 1 {
		return errors.New("ageStart < 1")
	}

	//ageEnd > 100
	if ad.Conditions.AgeEnd != 0 && ad.Conditions.AgeEnd > 100 {
		return errors.New("ageEnd > 100")
	}

	//Countries not in ISO-3166
	if len(ad.Conditions.Countries) != 0 {
		for _, country := range ad.Conditions.Countries {
			if !isISO3166(country) {
				return errors.New("Country not in ISO3166")
			}
		}
	}

	//Gender not in [M,F]
	if len(ad.Conditions.Gender) != 0 {
		for _, gender := range ad.Conditions.Gender {
			if !isValidGender(gender) {
				return errors.New("Gender can only be M or F")
			}
		}
	}

	//Platforms not in [android,ios,web]
	if len(ad.Conditions.Platforms) != 0 {
		for _, platform := range ad.Conditions.Platforms {
			if !isValidPlatform(platform) {
				return errors.New("Platform can only be android or web or ios")
			}
		}
	}

	return nil
}

func validateSearchParamAndAssignDefaultVal(r *http.Request) (SearchCondition, error) {

	var condition SearchCondition

	//Single value params
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return condition, errors.New("Offset value is invalid")
		}
		condition.Offset = offset
	} else {
		condition.Offset = 0
	}

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return condition, errors.New("Limit value is invalid")
		}
		condition.Limit = limit
	} else {
		condition.Limit = 5
	}

	for _, ageStr := range r.URL.Query()["age"] {
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			return condition, errors.New("Age value is invalid")
		}
		if age > 100 || age < 1 {
			return condition, errors.New("Age value is invalid")
		}
		condition.Age = append(condition.Age, ageStr)
	}

	//Multiple values params
	for _, value := range r.URL.Query()["gender"] {
		if !isValidGender(value) {
			return condition, errors.New("Gender value is invalid")
		}
		condition.Gender = append(condition.Gender, value)
	}

	for _, value := range r.URL.Query()["country"] {
		if !isISO3166(value) {
			return condition, errors.New("Country value is invalid")
		}
		condition.Country = append(condition.Country, value)
	}

	for _, value := range r.URL.Query()["platform"] {
		if !isValidPlatform(value) {
			return condition, errors.New("Platform value is invalid")
		}
		condition.Platform = append(condition.Platform, value)
	}

	return condition, nil
}

func isValidGender(gender string) bool {
	validGender := map[string]bool{
		"M": true,
		"F": true,
	}
	return validGender[gender]
}

func isValidPlatform(platform string) bool {
	validPlatform := map[string]bool{
		"ios":     true,
		"android": true,
		"web":     true,
	}
	return validPlatform[platform]
}

func isISO3166(country string) bool {
	iso3166 := map[string]bool{
		"AF": true, "AX": true, "AL": true, "DZ": true, "AS": true, "AD": true, "AO": true, "AI": true, "AQ": true, "AG": true, "AR": true, "AM": true, "AW": true, "AU": true, "AT": true, "AZ": true, "BS": true, "BH": true, "BD": true, "BB": true, "BY": true, "BE": true, "BZ": true, "BJ": true, "BM": true, "BT": true, "BO": true, "BQ": true, "BA": true, "BW": true, "BV": true, "BR": true, "IO": true, "BN": true, "BG": true, "BF": true, "BI": true, "CV": true, "KH": true, "CM": true, "CA": true, "KY": true, "CF": true, "TD": true, "CL": true, "CN": true, "CX": true, "CC": true, "CO": true, "KM": true, "CG": true, "CD": true, "CK": true, "CR": true, "CI": true, "HR": true, "CU": true, "CW": true, "CY": true, "CZ": true, "DK": true, "DJ": true, "DM": true, "DO": true, "EC": true, "EG": true, "SV": true, "GQ": true, "ER": true, "EE": true, "SZ": true, "ET": true, "FK": true, "FO": true, "FJ": true, "FI": true, "FR": true, "GF": true, "PF": true, "TF": true, "GA": true, "GM": true, "GE": true, "DE": true, "GH": true, "GI": true, "GR": true, "GL": true, "GD": true, "GP": true, "GU": true, "GT": true, "GG": true, "GN": true, "GW": true, "GY": true, "HT": true, "HM": true, "VA": true, "HN": true, "HK": true, "HU": true, "IS": true, "IN": true, "ID": true, "IR": true, "IQ": true, "IE": true, "IM": true, "IL": true, "IT": true, "JM": true, "JP": true, "JE": true, "JO": true, "KZ": true, "KE": true, "KI": true, "KP": true, "KR": true, "KW": true, "KG": true, "LA": true, "LV": true, "LB": true, "LS": true, "LR": true, "LY": true, "LI": true, "LT": true, "LU": true, "MO": true, "MK": true, "MG": true, "MW": true, "MY": true, "MV": true, "ML": true, "MT": true, "MH": true, "MQ": true, "MR": true, "MU": true, "YT": true, "MX": true, "FM": true, "MD": true, "MC": true, "MN": true, "ME": true, "MS": true, "MA": true, "MZ": true, "MM": true, "NA": true, "NR": true, "NP": true, "NL": true, "NC": true, "NZ": true, "NI": true, "NE": true, "NG": true, "NU": true, "NF": true, "MP": true, "NO": true, "OM": true, "PK": true, "PW": true, "PS": true, "PA": true, "PG": true, "PY": true, "PE": true, "PH": true, "PN": true, "PL": true, "PT": true, "PR": true, "QA": true, "RE": true, "RO": true, "RU": true, "RW": true, "BL": true, "SH": true, "KN": true, "LC": true, "MF": true, "PM": true, "VC": true, "WS": true, "SM": true, "ST": true, "SA": true, "SN": true, "RS": true, "SC": true, "SL": true, "SG": true, "SX": true, "SK": true, "SI": true, "SB": true, "SO": true, "ZA": true, "GS": true, "SS": true, "ES": true, "LK": true, "SD": true, "SR": true, "SJ": true, "SE": true, "CH": true, "SY": true, "TW": true, "TJ": true, "TZ": true, "TH": true, "TL": true, "TG": true, "TK": true, "TO": true, "TT": true, "TN": true, "TR": true, "TM": true, "TC": true, "TV": true, "UG": true, "UA": true, "AE": true, "GB": true, "US": true, "UM": true, "UY": true, "UZ": true, "VU": true, "VE": true, "VN": true, "VG": true, "VI": true, "WF": true, "EH": true, "YE": true, "ZM": true, "ZW": true,
	}
	return iso3166[country]
}
