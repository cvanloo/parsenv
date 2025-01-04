package parsenv_test

import (
	"testing"
	"reflect"

	"github.com/cvanloo/parsenv"
)

type testConfig struct {
	foo string `cfg:"-"`
	bar string `cfg:"required"`
	baz string `cfg:"default=hello world"`
	zab string `cfg:"name=ZaB"`
	rab string `cfg:"name=RaB;default=goodnight moon"`
	oof string `cfg:"name=oOF;required"`
	uwa int
	wou int    `cfg:"name=wou"`
	eew float64
}

func TestParsenv(t *testing.T) {
	var myConfig testConfig
	expectedConfig := testConfig{
		foo: "",
		bar: "bar value",
		baz: "baz value",
		zab: "zab value",
		rab: "goodnight moon",
		oof: "oof value",
		uwa: 0,
		wou: 5,
		eew: 6.7,
	}

	t.Setenv("FOO", "foo value") // must be ignored
	t.Setenv("BAR", "bar value")
	t.Setenv("BAZ", "baz value")
	t.Setenv("ZaB", "zab value")
	t.Setenv("oOF", "oof value")
	t.Setenv("wou", "5")
	t.Setenv("EEW", "6.7")

	// should not panic
	if err := parsenv.Load(&myConfig); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(myConfig, expectedConfig) {
		t.Errorf("expected %#v, got: %#v", expectedConfig, myConfig)
	}
}

func TestParsenvMissingRequired(t *testing.T) {
	var myConfig testConfig
	expectedConfig := testConfig{
		foo: "",
		bar: "",
		baz: "hello world",
		zab: "",
		rab: "goodnight moon",
		oof: "",
		uwa: 0,
		wou: 0,
		eew: 0.0,
	}
	err := parsenv.Load(&myConfig)
	if err != nil {
		t.Error("expected non-nil error, got nil")
	}
	werr, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Error("expected error to implement Unwrap() []error, but it does not")
	}
	errs := werr.Unwrap()
	if len(errs) != 2 {
		t.Errorf("expected to get 2 errors, but got: %d", len(errs))
	}
	if !reflect.DeepEqual(myConfig, expectedConfig) {
		t.Errorf("expected %#v, got: %#v", expectedConfig, myConfig)
	}
}
