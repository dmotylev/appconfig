// Copyright 2013 Dmitry Motylev
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

// Package nutrition provides decoding of different sources based on user defined struct.
// Source is the stream of lines in 'key=value' form. Environment, file and raw io.Reader sources are supported.
// Package can decode types boolean, numeric, and string types as well as time.Duration and time.Time. The later could be customized with formats (default is time.UnixDate).
package nutrition

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

// Build Nutrition instance that decodes environment variables. Given prefix used for looking values in environment
func Env(prefix string) Nutrition {
	return newFeeder().Env(prefix)
}

// Build Nutrition instance that decodes given Reader
func Reader(reader io.Reader) Nutrition {
	return newFeeder().Reader(reader)
}

// Build Nutrition instance that decodes given File.
// Do nothing if file does not exists.
func File(filename string) Nutrition {
	return newFeeder().File(filename)
}

type Nutrition interface {
	// Feeds given struct with decoded values from provided sources
	Feed(v interface{}) error
	// Add environment as source for values
	Env(prefix string) Nutrition
	// Add Reader as source for values
	Reader(reader io.Reader) Nutrition
	// Add File as source for values
	File(filename string) Nutrition
}

type NutritionError struct {
	// Raw value as it was read from source
	Value string
	// Name of the user struct
	Struct string
	// Name of the field in the struct
	Field string
	// Type of the field
	FieldType string
	// Cause why the field was not feed with value
	Cause error
}

func (e *NutritionError) Error() string {
	return fmt.Sprintf("nutrition.Feed: can't assign '%s' to %s.%s %s: %v", e.Value, e.Struct, e.Field, e.FieldType, e.Cause)
}

type context interface {
	err() error
	lookup(s reflect.StructField) (string, bool)
}

type harvester struct {
	contexts []context
}

func newFeeder() *harvester {
	return &harvester{make([]context, 0, 1)}
}

func (this *harvester) lookup(f reflect.StructField) (string, bool) {
	for _, ctx := range this.contexts {
		if v, found := ctx.lookup(f); found {
			return v, true
		}
	}
	return "", false
}

func (this *harvester) Feed(v interface{}) error {
	if reflect.Indirect(reflect.ValueOf(v)).Kind() != reflect.Struct {
		panic("nutrition: Feed for non-struct type")
	}
	for _, ctx := range this.contexts {
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
			continue
		}
		switch f.Kind() {
		case reflect.String:
			f.SetString(vString)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vUint, err := strconv.ParseUint(vString, 10, f.Type().Bits())
			if err != nil {
				return &NutritionError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
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
				return &NutritionError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
			}
			f.SetInt(vInt)
		case reflect.Float32, reflect.Float64:
			vFloat, err := strconv.ParseFloat(vString, f.Type().Bits())
			if err != nil {
				return &NutritionError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
			}
			f.SetFloat(vFloat)
		case reflect.Bool:
			vBool, err := strconv.ParseBool(vString)
			if err != nil {
				return &NutritionError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
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
				return &NutritionError{vString, vStruct.Type().Name(), fStruct.Name, f.Kind().String(), err}
			}
			f.Set(reflect.ValueOf(d))
		}
	}
	return nil
}

func (this *harvester) Env(prefix string) Nutrition {
	this.contexts = append(this.contexts, &envContext{prefix})
	return this
}

func (this *harvester) Reader(reader io.Reader) Nutrition {
	keyValue := regexp.MustCompile("^([^=]+)=(.*)$")
	ctx := &mapContext{nil, make(map[string]string)}
	this.contexts = append(this.contexts, ctx)
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

func (this *harvester) File(filename string) Nutrition {
	file, err := os.Open(filename)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		this.contexts = append(this.contexts, &mapContext{nil, make(map[string]string)})
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
