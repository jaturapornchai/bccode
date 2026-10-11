package report

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxCodeRunes = 50
	maxCodeList  = 200
)

// ResolveParams validates the browser parameters of one variant and returns the values the source runs with:
// defaults, then input, then the variant preset. Unknown or omitted parameters are invalid; a preset parameter
// may only be sent with its preset value (anything else = REPORT_VARIANT_FORBIDDEN)
func ResolveParams(d *Definition, v *Variant, input map[string]string) (map[string]string, *Failure) {
	byKey := map[string]Param{}
	for _, p := range d.Params {
		byKey[p.Key] = p
	}
	values := map[string]string{}
	for _, p := range offered(d, v) {
		values[p.Key] = p.Default
	}
	for key, raw := range input {
		if preset, locked := v.Preset[key]; locked {
			if strings.TrimSpace(raw) != preset {
				return nil, fail("REPORT_VARIANT_FORBIDDEN", key)
			}
			continue
		}
		if _, known := byKey[key]; !known || contains(v.Omit, key) {
			return nil, fail("REPORT_PARAM_INVALID", key)
		}
		values[key] = strings.TrimSpace(raw)
	}
	for key, preset := range v.Preset {
		values[key] = preset
	}
	for _, p := range d.Params {
		value, present := values[p.Key]
		if !present {
			continue
		}
		normalized, failure := checkParam(p, value)
		if failure != nil {
			return nil, failure
		}
		values[p.Key] = normalized
	}
	if failure := checkRanges(d, values); failure != nil {
		return nil, failure
	}
	return values, nil
}

func checkParam(p Param, value string) (string, *Failure) {
	if value == "" {
		if p.Required {
			return "", fail("REPORT_PARAM_REQUIRED", p.Key)
		}
		return "", nil
	}
	switch p.Type {
	case ParamDate:
		if !validDate(value) {
			return "", fail("REPORT_PARAM_INVALID", p.Key)
		}
	case ParamCode, ParamYear:
		if !validCode(value) {
			return "", fail("REPORT_PARAM_INVALID", p.Key)
		}
	case ParamCodes:
		codes := []string{}
		for _, part := range strings.Split(value, ",") {
			if part = strings.TrimSpace(part); part == "" {
				continue
			}
			if !validCode(part) {
				return "", fail("REPORT_PARAM_INVALID", p.Key)
			}
			codes = append(codes, part)
		}
		if len(codes) > maxCodeList {
			return "", fail("REPORT_PARAM_INVALID", p.Key)
		}
		return strings.Join(codes, ","), nil
	case ParamEnum:
		for _, o := range p.Options {
			if o.Value == value {
				return value, nil
			}
		}
		return "", fail("REPORT_PARAM_INVALID", p.Key)
	case ParamBool:
		if value != "true" && value != "false" {
			return "", fail("REPORT_PARAM_INVALID", p.Key)
		}
	default:
		return "", fail("REPORT_PARAM_INVALID", p.Key)
	}
	return value, nil
}

// checkRanges: date_from ≤ date_to and every <x>_from ≤ <x>_to code pair (byte order like the GL engine's COLLATE "C")
func checkRanges(d *Definition, values map[string]string) *Failure {
	if from, to := values["date_from"], values["date_to"]; from != "" && to != "" && from > to {
		return fail("REPORT_DATE_RANGE_INVALID", "date_from")
	}
	for _, p := range d.Params {
		if p.Type != ParamCode || !strings.HasSuffix(p.Key, "_from") {
			continue
		}
		toKey := strings.TrimSuffix(p.Key, "_from") + "_to"
		if from, to := values[p.Key], values[toKey]; from != "" && to != "" && from > to {
			return fail("REPORT_PARAM_INVALID", p.Key)
		}
	}
	return nil
}

func validDate(value string) bool {
	if len(value) != len("2006-01-02") {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validCode(value string) bool {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxCodeRunes {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || r == ',' {
			return false
		}
	}
	return true
}

// checkSort returns the sort to run: the default when empty, otherwise one of the definition's sorts
func checkSort(d *Definition, sort string) (string, *Failure) {
	if sort == "" {
		return d.DefaultSort, nil
	}
	for _, o := range d.Sorts {
		if o.Value == sort {
			return sort, nil
		}
	}
	return "", fail("REPORT_SORT_INVALID", "sort")
}
