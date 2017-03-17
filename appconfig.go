// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Source provides values for the fields
type Source interface {
	// Err returns the first error that was encountered by the source
	Err() error
	// Lookup returns string form value for the field and false if value ws not found
	Lookup(reflect.StructField) (string, bool)
}

// Error describes why it was impossible to assign a value to the struct's field.
type Error struct {
	typ   string
	fld   string
	kind  string
	val   string
	cause error
}

func (e *Error) Error() string {
	return fmt.Sprintf("can't set %s.%s (%s) to '%s': %v",
		e.typ, e.fld, e.kind, e.val, e.cause)
}

// Load populates given struct from ordered list of sources.
// First "positive" lookup will be used as a value for the field.
func Load(v interface{}, s ...Source) error {
	if reflect.Indirect(reflect.ValueOf(v)).Kind() != reflect.Struct {
		panic("scan for non-struct type")
	}
	if len(s) == 0 {
		return nil
	}
	scn := &scanner{s}
	return scn.scan(v)
}

type scanner struct {
	sources []Source
}

func (s *scanner) lookup(f reflect.StructField) (string, bool) {
	for _, src := range s.sources {
		if v, found := src.Lookup(f); found {
			return v, true
		}
	}
	return "", false
}

func (s *scanner) scan(v interface{}) error {
	for _, src := range s.sources {
		if err := src.Err(); err != nil {
			return err
		}
	}
	vStruct := reflect.ValueOf(v).Elem()
	for i := 0; i < vStruct.NumField(); i++ {
		f := vStruct.Field(i)
		if !f.CanSet() {
			continue
		}
		fStruct := vStruct.Type().Field(i)
		vString, found := s.lookup(fStruct)
		if !found {
			vString = fStruct.Tag.Get("default")
			if vString == "" {
				continue
			}
		}
		switch f.Kind() {
		case reflect.String:
			f.SetString(vString)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vUint, err := strconv.ParseUint(vString, 10, f.Type().Bits())
			if err != nil {
				return &Error{vStruct.Type().Name(), fStruct.Name, f.Kind().String(), vString, err}
			}
			f.SetUint(vUint)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var vInt int64
			var err error
			if f.Type().Name() == "Duration" { // time.Duration
				var d time.Duration
				d, err = time.ParseDuration(vString)
				vInt = int64(d)
			} else {
				vInt, err = strconv.ParseInt(vString, 10, f.Type().Bits())
			}
			if err != nil {
				return &Error{vStruct.Type().Name(), fStruct.Name, f.Kind().String(), vString, err}
			}
			f.SetInt(vInt)
		case reflect.Float32, reflect.Float64:
			vFloat, err := strconv.ParseFloat(vString, f.Type().Bits())
			if err != nil {
				return &Error{vStruct.Type().Name(), fStruct.Name, f.Kind().String(), vString, err}
			}
			f.SetFloat(vFloat)
		case reflect.Bool:
			vBool, err := strconv.ParseBool(vString)
			if err != nil {
				return &Error{vStruct.Type().Name(), fStruct.Name, f.Kind().String(), vString, err}
			}
			f.SetBool(vBool)
		case reflect.Struct:
			if f.Type().Name() != "Time" { // time.Time
				continue
			}
			format := fStruct.Tag.Get("time,format")
			if format == "" {
				format = time.UnixDate
			}
			d, err := time.Parse(format, vString)
			if err != nil {
				return &Error{vStruct.Type().Name(), fStruct.Name, f.Kind().String(), vString, err}
			}
			f.Set(reflect.ValueOf(d))
		}
	}
	return nil
}
