// json is an attempt to provide a drop-in replacement of encoding/json that
// makes it trivial to add some schema-style validation to json decoding.
//
// As a minimal proof that this is possible I've implemented only a single
// layer of enforcement through an Options struct which may be passed into
// UnmarshalX. As an example:
//
//   err := json.Unmarshal(data, &dest, json.Options{
//     NullNotPresent: []string{"host", "user"},
//     Required:       []string{"host", "user"},
//   })
//
// would produce an error if host or user was unset or was set to 'null'.
package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Options struct {
	// Pedantic will treat every any field that isn't specified in the struct as
	// forbidden and ever field that is specified in the struct as required. If
	// Pedantic is set it does not impact how null is treated, that is still
	// driven through GlobalNullNotPresent or NullNotPresent.
	//
	// Currently unsupported. :(
	Pedantic bool // TODO

	// Strict will treat any key that not specified in the destination struct as
	// a forbidden key.
	//
	// Currently unsupported. :(
	Strict bool // TODO

	// GlobalNullNotPresent will force UnmarshalX to act as if NullNotPresent is
	// set for every key.
	GlobalNullNotPresent bool

	// FailFast will abort unmarshalling on the first encountered error.
	FailFast bool

	// NullNotPresent is a set of keys that will treat null as an unset value.
	// The default behavior is that `"key": null` will satisfy key presence for
	// the Required keys. For keys contained in this set an error would be thrown
	// instead.
	NullNotPresent []string

	// Required is a set of keys that must be set in the json being unmarshalled.
	// Any Unmarshal of json containing these keys will return an error if they
	// are not present.
	Required []string

	// Forbidden specifies a set of keys that must *not* be set in the json being
	// unmarshalled. If they are present they will result in an error.
	Forbidden []string
}

type builtOptions struct {
	Options
	nullNotPresentSet map[string]bool
}

func prepareOptions(o Options, v interface{}) builtOptions {
	bo := builtOptions{o, map[string]bool{}}
	for _, k := range bo.NullNotPresent {
		bo.nullNotPresentSet[k] = true
	}

	return bo
}

// given a key return if null should be considered a "set" value
func (bo builtOptions) nullIsPresent(s string) bool {
	if bo.GlobalNullNotPresent {
		return false
	}

	return !bo.nullNotPresentSet[s]
}

func Unmarshal(data []byte, v interface{}) error {
	// defer unmarshal calls to UnmarshalX; this will eventually let us take over
	// unmarshalling of nested structs if appropriate as well as act on struct
	// tag annotations.
	return UnmarshalX(data, v, nil)
}

// UnmarshalX reads json from data and stores keys into v while enforcing any
// Options that were passed in. Passing pcfg as nil will result in no options
// being applied (i.e. it behaves as json.Unmarshal).
func UnmarshalX(data []byte, v interface{}, pcfg *Options) error {
	if pcfg == nil {
		// eventually we'll still need to UnmarshalX on the children in case they
		// have options configured
		return json.Unmarshal(data, v)
	}

	dest := make(map[string]*json.RawMessage)
	err := json.Unmarshal(data, &dest)
	if err != nil {
		return err
	}

	// build interal state
	cfg := prepareOptions(*pcfg, v)

	errors := []ValidationError{}
	addError := func(ve ValidationError) bool {
		errors = append(errors, ve)
		return cfg.FailFast
	}

	present := func(s string) bool {
		v, ok := dest[s]
		switch {
		case !ok:
			return false
		case v == nil && cfg.GlobalNullNotPresent:
			return false
		case v == nil && !cfg.nullIsPresent(s):
			return false
		}
		return true
	}

	for _, reqKey := range cfg.Required {
		if !present(reqKey) && addError(ValidationError{MissingKey, reqKey}) {
			goto done
		}
	}

	for _, forbKey := range cfg.Forbidden {
		if addError(ValidationError{ForbiddenKey, forbKey}) {
			goto done
		}
	}

done:
	if len(errors) != 0 {
		return ErrorCollection{errors}
	} else {
		return json.Unmarshal(data, v)
	}
	return nil
}

// -- Error types --

// ErrorCollection is a set of errors that were encountered when enforcing the
// requested unmarshal options.
type ErrorCollection struct {
	errors []ValidationError
}

var _ error = ErrorCollection{}

func (e ErrorCollection) Error() string {
	s := make([]string, len(e.errors))
	for i, ele := range e.errors {
		s[i] = ele.Error()
	}
	return fmt.Sprintf("['%s']", strings.Join(s, "', '"))
}

// ValidationErrorType specifies which type of validation error was encountered
type ValidationErrorType int

const (
	MissingKey ValidationErrorType = iota
	ForbiddenKey
)

// ValidationError is a binds together a ValidationErrorType and the key that
// failed to validate in the appropriate way.
type ValidationError struct {
	Type ValidationErrorType
	Key  string
}

var _ error = ValidationError{}

func (ve ValidationError) Error() string {
	if ve.Type == MissingKey {
		return missingKey(ve.Key)
	}
	if ve.Type == ForbiddenKey {
		return forbiddenKey(ve.Key)
	}

	return fmt.Sprintf("unexpected error type %d for key <%s>", ve.Type, ve.Key)
}

func missingKey(s string) string {
	return fmt.Sprintf("required key <%s> not found", s)
}

func forbiddenKey(s string) string {
	return fmt.Sprintf("forbidden key <%s> was set", s)
}

// -- defer everything except unmarshal to the default library --

func Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

func HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}
