package nutrition

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Env(prefix string) Nutrition {
	return newHarvester().Env(prefix)
}

func File(filename string) Nutrition {
	return newHarvester().File(filename)
}

type Nutrition interface {
	Harvest(v interface{}) error
	Env(prefix string) Nutrition
	File(filename string) Nutrition
}

type NutritionError struct {
	Value     string
	Struct    string
	Field     string
	FieldType string
	Cause     error
}

func (e *NutritionError) Error() string {
	return fmt.Sprintf("nutrition.Harvest: can't assign '%s' to %s.%s %s: %v", e.Value, e.Struct, e.Field, e.FieldType, e.Cause)
}

type context interface {
	lookup(s reflect.StructField) (string, bool)
}

type harvester struct {
	contexts []context
}

func newHarvester() *harvester {
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

func (this *harvester) Harvest(v interface{}) error {
	if reflect.Indirect(reflect.ValueOf(v)).Kind() != reflect.Struct {
		panic("nutrition: Harvest for non-struct type")
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

func (this *harvester) File(filename string) Nutrition {
	this.contexts = append(this.contexts, &fileContext{filename})
	return this
}

// environment context

type envContext struct {
	prefix string
}

func (this *envContext) lookup(s reflect.StructField) (string, bool) {
	k := strings.ToUpper(this.prefix + s.Name)
	v := os.Getenv(k)
	return v, v != ""
}

// file context

type fileContext struct {
	filename string
}

func (this *fileContext) lookup(s reflect.StructField) (string, bool) {
	return "", false
}
