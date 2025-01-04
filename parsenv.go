package parsenv

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
	"unsafe"
	"errors"
)

// @todo: support ints and floats aside from strings
// @todo: infer name, camel case to SCREAMING_SNAKE_CASE

type TagData struct {
	Name string
	Default string
	Required bool
	Ignored bool
}

func Load(cfg any) (err error) {
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
				if td.Ignored {
					// ignore
				} else if strVal := os.Getenv(optionName); strVal != "" {
					setUnexportedField(val, strVal)
				} else if td.Default != "" {
					setUnexportedField(val, td.Default)
				} else if td.Required {
					err = errors.Join(err, fmt.Errorf("missing env value for required field: %s", field.Name))
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
		switch len(propertyParts) {
		default:
			panic(fmt.Sprintf("invalid format for property in cfg tag: %s", rawProperty)) // @todo: better error message (location?)
		case 1:
			switch propertyParts[0] {
			default:
			case "-":
				td.Ignored = true
			case "required":
				td.Required = true
			}
		case 2:
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

func setUnexportedField(field reflect.Value, value any) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}
