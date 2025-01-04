// The parsenv package exposes a Load function that populates the fields of a
// struct with data from environment variables.
//
//   var myConfig struct {
//   	foo string  `cfg:"required"`
//   	bar int     `cfg:"default=15"`
//   	baz float64 `cfg:"name=bAz;default=6.97"`
//   }
//   
//   if err := parsenv.Load(&myConfig); err != nil {
//   	log.Fatal(err)
//   }
//
// Per default, field names are converted from PascalCase or camelCase to
// SCREAMING_SNAKE_CASE.
//
// For parsing options refer to the documentation of parsenv.TagData.
package parsenv

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

// The behavior of how the environment is read into a struct can be influenced
// with the `cfg` struct tag.
//
//   var myConfig struct{
//   	foo int     `cfg:"-"`                    // this field is ignored
//   	bar float64 `cfg:"required"`             // return an error if BAR is not found in the environment
//   	baz string  `cfg:"name=baz"`             // specify a custom name for the env var (per default the field name is converted to SCREAMING_SNAKE_CASE)
//   	zap string  `cfg:"default=hello world"`  // specify a default value
//   	puf int     `cfg:"name=PUFF;default=19"` // use ; to specify multiple properties
//   }
type TagData struct {
	Name     string // name=<name>
	Default  string // default=<value>
	Required bool   // required
	Ignored  bool   // -
}

// Load reads environment variables into a struct.
// If the cfg variable passed is not a pointer to a struct, Load will panic.
// If any of the fields contain invalid `cfg` struct tags, Load will panic also.
// If one or more fields marked as 'required' don't have a corresponding
// environment variable, Load will return an error.
func Load(cfg any) (err error) {
	cfgRefl := reflect.ValueOf(cfg)
	cfgType := cfgRefl.Type()
	if cfgType.Kind() != reflect.Pointer {
		panic("parsenv.Load: must pass a pointer")
	}
	for _, field := range reflect.VisibleFields(cfgType.Elem()) {
		optionName := changeNameCase(field.Name)
		td := parseTag(field.Tag.Get("cfg"))
		if td.Name != "" {
			optionName = td.Name
		}
		if val := cfgRefl.Elem().Field(field.Index[0]); val.IsValid() {
			if td.Ignored {
				// ignore
			} else if strVal := os.Getenv(optionName); strVal != "" {
				optVal, perr := parseValue(field.Type.Kind(), strVal)
				err = errors.Join(err, perr)
				setUnexportedField(val, optVal)
			} else if td.Default != "" {
				optVal, perr := parseValue(field.Type.Kind(), td.Default)
				err = errors.Join(err, perr)
				setUnexportedField(val, optVal)
			} else if td.Required {
				err = errors.Join(err, fmt.Errorf("missing env value for required field: %s", field.Name))
			}
		}
	}
	return err
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

func parseValue(kind reflect.Kind, val string) (any, error) {
	switch kind {
	default:
		panic("only the types string, int, and float64 are supported")
	case reflect.String:
		return val, nil
	case reflect.Int:
		return strconv.Atoi(val)
	case reflect.Float64:
		return strconv.ParseFloat(val, 64)
	}
}

func setUnexportedField(field reflect.Value, value any) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}
