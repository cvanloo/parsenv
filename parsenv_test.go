package parsenv

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func ExampleLoad() {
	os.Setenv("FOO", "こんにちは、世界！")
	os.Setenv("BAZ", "13.37")
	os.Setenv("QUX", "yes")

	var myConfig struct {
		foo string  `cfg:"required"`
		bar int     `cfg:"default=15"`
		baz float64 `cfg:"name=bAz;default=6.97"`
		qux bool
	}

	if err := Load(&myConfig); err != nil {
		log.Fatal(err)
	}

	// because BAZ does not match the custom name bAz, the default value is applied.
	fmt.Println(myConfig.foo, myConfig.bar, myConfig.baz, myConfig.qux)
	// Output: こんにちは、世界！ 15 6.97 true
}

type testConfig struct {
	foo string `cfg:"-"`
	bar string `cfg:"required"`
	baz string `cfg:"default=hello world"`
	zab bool   `cfg:"name=ZaB"`
	rab string `cfg:"name=RaB;default=goodnight moon"`
	oof string `cfg:"name=oOF;required"`
	uwa int
	wou int `cfg:"name=wou"`
	eew float64
}

func TestLoad(t *testing.T) {
	var myConfig testConfig
	expectedConfig := testConfig{
		foo: "",
		bar: "bar value",
		baz: "baz value",
		zab: true,
		rab: "goodnight moon",
		oof: "oof value",
		uwa: 0,
		wou: 5,
		eew: 6.7,
	}

	t.Setenv("FOO", "foo value") // must be ignored
	t.Setenv("BAR", "bar value")
	t.Setenv("BAZ", "baz value")
	t.Setenv("ZaB", "YES")
	t.Setenv("oOF", "oof value")
	t.Setenv("wou", "5")
	t.Setenv("EEW", "6.7")

	// should not panic
	if err := Load(&myConfig); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(myConfig, expectedConfig) {
		t.Errorf("expected %#v, got: %#v", expectedConfig, myConfig)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	var myConfig testConfig
	expectedConfig := testConfig{
		foo: "",
		bar: "",
		baz: "hello world",
		zab: false,
		rab: "goodnight moon",
		oof: "",
		uwa: 0,
		wou: 0,
		eew: 0.0,
	}
	err := Load(&myConfig)
	if err == nil {
		t.Error("expected non-nil error, got nil")
	}
	werr, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Error("expected error to implement Unwrap() []error, but it does not")
	} else {
		errs := werr.Unwrap()
		if len(errs) != 2 {
			t.Errorf("expected to get 2 errors, but got: %d", len(errs))
		}
	}
	if !reflect.DeepEqual(myConfig, expectedConfig) {
		t.Errorf("expected %#v, got: %#v", expectedConfig, myConfig)
	}
}

func TestLoadNotAPointer(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected a panic, got nothing")
		}
	}()
	var myConfig struct{}
	Load(myConfig)
}

func TestLoadNotAStruct(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected a panic, got nothing")
		}
	}()
	var myConfig int
	Load(&myConfig)
}

func TestCaseChange(t *testing.T) {
	c1 := changeNameCase("helloGoodWorld")
	if c1 != "HELLO_GOOD_WORLD" {
		t.Errorf("expected HELLO_GOOD_WORLD, got: %s", c1)
	}
	c2 := changeNameCase("HelloGentleMoon")
	if c2 != "HELLO_GENTLE_MOON" {
		t.Errorf("expected HELLO_GENTLE_MOON, got: %s", c2)
	}
	c3 := changeNameCase("someoneReallyLikesACRONYMS")
	if c3 != "SOMEONE_REALLY_LIKES_ACRONYMS" {
		t.Errorf("expected SOMEONE_REALLY_LIKES_ACRONYMS, got: %s", c3)
	}
}
