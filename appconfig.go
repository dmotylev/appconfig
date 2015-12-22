// Copyright 2015 Dmitry Motylev
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package appconfig provides initialization for user defined struct
// from configured sources. Source is the stream of lines in 'key=value' form.
// OS environment, file and raw io.Reader sources are supported. The package
// can decode boolean, numeric, string, time.Duration and time.Time.
// The later could be customized with formats
package appconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Build AppConfig instance that decodes environment variables.
// Given prefix used for looking values in environment
func Env(prefix string) AppConfig {
	return newScanner().Env(prefix)
}

// Build AppConfig instance that decodes given Reader
func Reader(reader io.Reader) AppConfig {
	return newScanner().Reader(reader)
}

// Build AppConfig instance that decodes given File.
// Do nothing if file does not exists.
func File(filename string) AppConfig {
	return newScanner().File(filename)
}

type AppConfig interface {
	// Scans given struct with decoded values from provided sources
	Scan(v interface{}) error
	// Add environment as source for values
	Env(prefix string) AppConfig
	// Add Reader as source for values
	Reader(reader io.Reader) AppConfig
	// Add File as source for values
	File(filename string) AppConfig
}

type AppConfigError struct {
	// Raw value as it was read from source
	Value string
	// Name of the user struct
	Struct string
	// Name of the field in the struct
	Field string
	// Type of the field
	FieldType string
	// Cause why value was not assigned to the field
	Cause error
}

func (e *AppConfigError) Error() string {
	return fmt.Sprintf("appconfig: can't assign '%s' to %s.%s %s: %v",
		e.Value, e.Struct, e.Field, e.FieldType, e.Cause)
}

type source interface {
	err() error
	lookup(s reflect.StructField) (string, bool)
}

type scanner struct {
	sources []source
}

func newScanner() *scanner {
	return &scanner{make([]source, 0, 1)}
}

func (this *scanner) lookup(f reflect.StructField) (string, bool) {
	for _, ctx := range this.sources {
		if v, found := ctx.lookup(f); found {
			return v, true
		}
	}
	return "", false
}

func (this *scanner) Scan(v interface{}) error {
	if reflect.Indirect(reflect.ValueOf(v)).Kind() != reflect.Struct {
		panic("appconfig: scan for non-struct type")
	}
	for _, ctx := range this.sources {
		if err := ctx.err(); err != nil {
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
		vString, found := this.lookup(fStruct)
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
				return &AppConfigError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
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
				return &AppConfigError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
			}
			f.SetInt(vInt)
		case reflect.Float32, reflect.Float64:
			vFloat, err := strconv.ParseFloat(vString, f.Type().Bits())
			if err != nil {
				return &AppConfigError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
			}
			f.SetFloat(vFloat)
		case reflect.Bool:
			vBool, err := strconv.ParseBool(vString)
			if err != nil {
				return &AppConfigError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
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
				return &AppConfigError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
			}
			f.Set(reflect.ValueOf(d))
		}
	}
	return nil
}

func (this *scanner) Env(prefix string) AppConfig {
	this.sources = append(this.sources, &envContext{prefix})
	return this
}

func (this *scanner) Reader(reader io.Reader) AppConfig {
	keyValue := regexp.MustCompile("^([^=]+)=(.*)$")
	ctx := &mapContext{nil, make(map[string]string)}
	this.sources = append(this.sources, ctx)
	scanner := bufio.NewScanner(bufio.NewReader(reader))
	for scanner.Scan() {
		line := scanner.Text()
		parts := keyValue.FindStringSubmatch(line)
		if parts == nil {
			continue
		}
		ctx.values[strings.ToLower(parts[1])] = parts[2]
	}
	if err := scanner.Err(); err != nil {
		ctx.errr = err
	}
	return this
}

func (this *scanner) File(filename string) AppConfig {
	file, err := os.Open(filename)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		this.sources = append(this.sources, &mapContext{nil, make(map[string]string)})
		return this
	}
	return this.Reader(file)
}

// environment context

type envContext struct {
	prefix string
}

func (this *envContext) err() error {
	return nil
}

func (this *envContext) lookup(s reflect.StructField) (string, bool) {
	k := strings.ToUpper(this.prefix + s.Name)
	v := os.Getenv(k)
	return v, v != ""
}

// map context

type mapContext struct {
	errr   error
	values map[string]string
}

func (this *mapContext) err() error {
	return this.errr
}

func (this *mapContext) lookup(s reflect.StructField) (string, bool) {
	v, found := this.values[strings.ToLower(s.Name)]
	return v, found
}
