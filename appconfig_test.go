// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"testing/iotest"
	"time"

	"github.com/dmotylev/appconfig"
)

type DummyStruct struct{}

type Fields struct {
	String     string
	Uint       uint64
	Int        int64
	Float      float64
	Bool       bool
	Duration   time.Duration
	Date       time.Time `time,format:"2006-01-02T15:04:05Z07:00"`
	DateUnix   time.Time
	Struct     DummyStruct
	nonSetable int
}

const stream = `
string=123.4
uint=123
int=-123
float=123.4
bool=true
duration=1h2m3s
date=2006-01-02T15:04:05+07:00
dateunix=Mon Jan 2 15:04:05 MST 2006
struct=Dummy
nonsetable=123
novalue=
onlykey
`

func test(f Fields, t *testing.T) {
	if f.nonSetable != 0 {
		t.Errorf("f.nonSetable = %v, want 0", f.nonSetable)
	}
	if f.String != "123.4" {
		t.Errorf("f.String = '%v', want '123.4'", f.String)
	}
	if f.Uint != 123 {
		t.Errorf("f.Uint = %v, want 123", f.Uint)
	}
	if f.Int != -123 {
		t.Errorf("f.Int = %v, want -123", f.Int)
	}
	if f.Float != float64(123.4) {
		t.Errorf("f.Float = %v, want 123.4", f.Float)
	}
	if !f.Bool {
		t.Errorf("f.Bool = %v, want true", f.Bool)
	}
	date, e := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	if e != nil {
		t.Fatal(e)
	}
	if !f.Date.Equal(date) {
		t.Errorf("f.Date = %v, want %v", f.Date, date)
	}
	dur, _ := time.ParseDuration("1h2m3s")
	if f.Duration != dur {
		t.Errorf("f.Duration = %v, want 1h2m3s", f.Duration)
	}
}

func TestLoad_NonStruct(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic on non-struct type, want panic")
		}
	}()
	var i int
	appconfig.Load(i)
}

func TestLoad_NoSources(t *testing.T) {
	f := Fields{
		String:   "123.4",
		Uint:     123,
		Int:      -123,
		Float:    123.4,
		Bool:     true,
		Duration: (time.Duration(1) * time.Hour) + (time.Duration(2) * time.Minute) + (time.Duration(3) * time.Second),
		Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", int(7*int64(time.Hour/time.Second)))),
	}

	if err := appconfig.Load(&f); err != nil {
		t.Error(err)
	}

	test(f, t)
}

func TestLoad_FromEnvWithDefault(t *testing.T) {
	os.Clearenv()

	const s = `owls are not what do you think about them`

	var f struct {
		String string `default:"owls are not what do you think about them"`
	}

	if err := appconfig.Load(&f, appconfig.FromEnv("app_")); err != nil {
		t.Error(err)
	}
	if f.String != s {
		t.Errorf("f.String == '%v', want '%s'", f.String, s)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_STRING", "123.4")
	os.Setenv("APP_UINT", "123")
	os.Setenv("APP_INT", "-123")
	os.Setenv("APP_FLOAT", "123.4")
	os.Setenv("APP_BOOL", "true")
	os.Setenv("APP_DURATION", "1h2m3s")
	os.Setenv("APP_DATE", "2006-01-02T15:04:05+07:00")
	os.Setenv("APP_DATEUNIX", "Mon Jan 2 15:04:05 MST 2006")
	os.Setenv("APP_STRUCT", "Dummy")
	os.Setenv("APP_NONSETABLE", "123")

	var f Fields
	if err := appconfig.Load(&f, appconfig.FromEnv("app_")); err != nil {
		t.Error(err)
	}

	test(f, t)
}

func testLoad_FromEnvInvSrc(k string, t *testing.T) {
	os.Clearenv()
	os.Setenv(k, "Err")

	var f Fields
	err := appconfig.Load(&f, appconfig.FromEnv("app_"))
	if err == nil {
		t.Errorf("err==nil for invalid %s, want not nil", k)
	}
	if err.Error() == "" {
		t.Error(`err.Error()=="", want non empty string`)
	}
}

func TestLoad_FromEnvUintErr(t *testing.T) {
	testLoad_FromEnvInvSrc("APP_UINT", t)
}

func TestLoad_FromEnvIntErr(t *testing.T) {
	testLoad_FromEnvInvSrc("APP_INT", t)
}

func TestLoad_FromEnvFloatErr(t *testing.T) {
	testLoad_FromEnvInvSrc("APP_FLOAT", t)
}

func TestLoad_FromEnvBoolErr(t *testing.T) {
	testLoad_FromEnvInvSrc("APP_BOOL", t)
}

func TestLoad_FromEnvDurationErr(t *testing.T) {
	testLoad_FromEnvInvSrc("APP_DURATION", t)
}

func TestLoad_FromEnvDateErr(t *testing.T) {
	testLoad_FromEnvInvSrc("APP_DATE", t)
}

func createFile(content string) (*os.File, error) {
	file, err := ioutil.TempFile(os.TempDir(), "appconfig_test")
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(file, content); err != nil {
		return nil, err
	}
	return file, err
}

func TestLoad_FromFile(t *testing.T) {
	file, err := createFile(stream)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	var f Fields
	if err := appconfig.Load(&f, appconfig.FromFile(file.Name())); err != nil {
		t.Fatal(err)
	}

	test(f, t)
}

func TestLoad_FromNoFile(t *testing.T) {
	var f Fields
	if err := appconfig.Load(&f, appconfig.FromFile(":$:")); err != nil {
		t.Errorf("err == '%s', want nil", err)
	}
}

func TestLoad_FromReader(t *testing.T) {
	var f Fields
	if err := appconfig.Load(&f, appconfig.FromReader(bytes.NewBufferString(stream))); err != nil {
		t.Error(err)
	}

	test(f, t)
}

func TestLoad_FromReaderError(t *testing.T) {
	var f Fields
	if appconfig.Load(&f, appconfig.FromReader(iotest.TimeoutReader(bytes.NewBufferString(stream)))) == nil {
		t.Errorf("err == nil, want not nil")
	}
}

func TestLoad_FromEnvFromReader(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_STRING", "123.4")
	reader := bytes.NewBufferString("int=-123")

	var f Fields
	if err := appconfig.Load(&f, appconfig.FromEnv("app_"), appconfig.FromReader(reader)); err != nil {
		t.Error(err)
	}

	if f.String != "123.4" {
		t.Errorf("f.String = '%v', want '123.4'", f.String)
	}
	if f.Int != -123 {
		t.Errorf("f.Int = %v, want -123", f.Int)
	}
}

func TestLoad_FromReaderFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_STRING", "123.4")
	reader := bytes.NewBufferString("int=-123")

	var f Fields
	if err := appconfig.Load(&f, appconfig.FromReader(reader), appconfig.FromEnv("app_")); err != nil {
		t.Error(err)
	}

	if f.String != "123.4" {
		t.Errorf("f.String = '%v', want '123.4'", f.String)
	}
	if f.Int != -123 {
		t.Errorf("f.Int = %v, want -123", f.Int)
	}
}

func TestErr(t *testing.T) {
	if v, found := appconfig.Err(nil).Lookup(reflect.StructField{}); v != "" || found {
		t.Error("false positive")
	}
}

func ExampleLoad_FromEnv() {
	os.Clearenv()
	os.Setenv("APP_TIMEOUT", "1h2m3s")
	os.Setenv("APP_DAY", "2013-12-13")
	os.Setenv("APP_NUMWORKERS", "5")
	os.Setenv("APP_WORKERNAME", "Hulk")

	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := appconfig.Load(&conf, appconfig.FromEnv("APP_"))

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
	// Output:
	// err=<nil>
	// timeout=1h2m3s
	// day=Fri Dec 13 00:00:00 UTC 2013
	// worker=Hulk
	// workers=5
}
