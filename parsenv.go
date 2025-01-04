package parsenv

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

// @todo: support ints and floats aside from strings
// @todo: infer name, camel case to SCREAMING_SNAKE_CASE
// @todo: required property
// @todo: ignore (-) property

type TagData struct {
	Name string
	Default string
}

func Load(cfg any) error {
	cfgRefl := reflect.ValueOf(cfg)
	cfgType := cfgRefl.Type()
	if cfgType.Kind() != reflect.Pointer {
		panic("parsenv.Load: must pass a pointer")
	}
	for _, field := range reflect.VisibleFields(cfgType.Elem()) {
		if field.Type.Kind() == reflect.String {
			optionName := changeNameCase(field.Name)
			td := parseTag(field.Tag.Get("cfg"))
			if td.Name != "" {
				optionName = td.Name
			}
			if val := cfgRefl.Elem().Field(field.Index[0]); val.IsValid() {
				if strVal := os.Getenv(optionName); strVal != "" {
					val.Set(reflect.ValueOf(strVal))
				} else if td.Default != "" {
					val.Set(reflect.ValueOf(td.Default))
				}
			}
		}
	}
	return nil
}

func parseTag(rawTag string) (td TagData) {
	if rawTag == "" {
		return td
	}
	rawParts := strings.Split(rawTag, ";")
	for _, rawProperty := range rawParts {
		propertyParts := strings.Split(rawProperty, "=")
		if len(propertyParts) != 2 {
			panic(fmt.Sprintf("invalid format for property in cfg tag: %s", rawProperty)) // @todo: better error message (location?)
		}
		key := propertyParts[0]
		val := propertyParts[1]
		switch key {
		default:
			panic(fmt.Sprintf("unknown property in cfg tag: %s", key))
		case "name":
			td.Name = val
		case "default":
			td.Default = val
		}
	}
	return td
}

func changeNameCase(name string) string {
	runes := []rune(name)
	caseChangeIdxs := []int{0}
	for i := range runes[1:] {
		if unicode.IsLower(runes[i]) && unicode.IsUpper(runes[i+1]) {
			caseChangeIdxs = append(caseChangeIdxs, i+1)
		}
	}
	var screamingSnakeCase strings.Builder
	for i := range caseChangeIdxs[1:] {
		s := caseChangeIdxs[i]
		e := caseChangeIdxs[i+1]
		for _, r := range runes[s:e] {
			screamingSnakeCase.WriteRune(unicode.ToUpper(r))
		}
		screamingSnakeCase.WriteRune('_')
	}
	for _, r := range runes[caseChangeIdxs[len(caseChangeIdxs)-1]:] {
		screamingSnakeCase.WriteRune(unicode.ToUpper(r))
	}
	return screamingSnakeCase.String()
}
