package json

import (
	"fmt"
	"reflect"
	"testing"
)

func init() {
	fmt.Printf("")
	i := 4444
	ts = TestStruct{"foo", &i}
}

type TestStruct struct {
	Foo string `json:"foo"`
	Bar *int   `json:"bar"`
}

var ts TestStruct

var tsEncoded = []byte(`{
  "foo": "foo",
  "bar": 4444
}`)

func testTS(t *testing.T, got, want TestStruct) {
	if got.Foo != want.Foo {
		t.Errorf("got: %v, want: %v", got.Foo, want.Foo)
	}

	switch {
	case got.Bar == nil && want.Bar == nil:
		return
	case got.Bar == nil && want.Bar != nil:
		t.Errorf("got: nil, want: %d", *want.Bar)
	case got.Bar != nil && want.Bar == nil:
		t.Errorf("got: %d, want: nil", *got.Bar)
	case *got.Bar != *want.Bar:
		t.Errorf("got: %d, want: %d", *got.Bar, *want.Bar)
	}
}

func noErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("got: %#v, want: nil", err)
	}
}

func TestUnmarshalXNoOpts(t *testing.T) {
	o := TestStruct{}
	err := UnmarshalX(tsEncoded, &o, &Options{})
	noErr(t, err)
	testTS(t, ts, o)
}

func TestUnmarshalXRejectBar(t *testing.T) {
	o := TestStruct{}
	e := UnmarshalX(tsEncoded, &o, &Options{Forbidden: []string{"bar"}})
	if e == nil {
		t.Errorf("got: nil, want: error")
		return
	}

	err, ok := e.(ErrorCollection)
	if !ok {
		t.Errorf("got: %T, %#v, want: ErrorCollection", e, e)
	}

	want := ErrorCollection{[]ValidationError{
		{ForbiddenKey, "bar"},
	}}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("got: %#v, want: %#v", err, want)
	}
}

func TestUnmarshalXRejectBarRequireFoo(t *testing.T) {
	input := []byte(`{"bar": 4444}`)
	o := TestStruct{}
	cfg := &Options{Required: []string{"foo"}, Forbidden: []string{"bar"}}

	e := UnmarshalX(input, &o, cfg)
	if e == nil {
		t.Errorf("got: nil, want: error")
		return
	}

	err, ok := e.(ErrorCollection)
	if !ok {
		t.Errorf("got: %T, %#v, want: ErrorCollection", e, e)
	}

	want := ErrorCollection{[]ValidationError{
		{MissingKey, "foo"},
		{ForbiddenKey, "bar"},
	}}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("got: %#v, want: %#v", err, want)
	}
}

func TestUnmarshalXRejectBarRequireFooFailFast(t *testing.T) {
	input := []byte(`{"bar": 4444}`)
	o := TestStruct{}
	cfg := &Options{FailFast: true, Required: []string{"foo"}, Forbidden: []string{"bar"}}

	e := UnmarshalX(input, &o, cfg)
	if e == nil {
		t.Errorf("got: nil, want: error")
		return
	}

	err, ok := e.(ErrorCollection)
	if !ok {
		t.Errorf("got: %T, %#v, want: ErrorCollection", e, e)
	}

	want := ErrorCollection{[]ValidationError{{MissingKey, "foo"}}}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("got: %#v, want: %#v", err, want)
	}
}

func TestUnmarshalXNullNotPresent(t *testing.T) {
	input := []byte(`{"foo": null,"bar": 4444}`)

	o := TestStruct{}
	cfg := &Options{
		Required:       []string{"foo"},
		NullNotPresent: []string{"foo"},
	}

	e := UnmarshalX(input, &o, cfg)
	err, ok := e.(ErrorCollection)

	if !ok {
		t.Errorf("got: %T, %#v, want: ErrorCollection", e, e)
		return
	}

	want := ErrorCollection{[]ValidationError{{MissingKey, "foo"}}}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("got: %#v, want: %#v", err, want)
	}
}

func TestUnmarshalXGlobalNullNotPresentForbiddenNoFailFast(t *testing.T) {
	input := []byte(`{"foo": null,"bar": 4444}`)

	o := TestStruct{}
	cfg := &Options{
		Forbidden:            []string{"bar"},
		Required:             []string{"foo"},
		GlobalNullNotPresent: true,
	}

	e := UnmarshalX(input, &o, cfg)
	err, ok := e.(ErrorCollection)

	if !ok {
		t.Errorf("got: %T, %#v, want: ErrorCollection", e, e)
		return
	}

	want := ErrorCollection{[]ValidationError{
		{MissingKey, "foo"},
		{ForbiddenKey, "bar"},
	}}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("got: %#v, want: %#v", err, want)
	}
}
