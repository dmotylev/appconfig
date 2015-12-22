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

package appconfig

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"testing/iotest"
	"time"
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

func verify(f Fields, t *testing.T) {
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

func TestScanPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic on non-struct type, want panic")
		}
	}()
	var i int
	(&scanner{}).Scan(i)
}

func TestScan(t *testing.T) {
	f := Fields{
		String:   "123.4",
		Uint:     123,
		Int:      -123,
		Float:    123.4,
		Bool:     true,
		Duration: (time.Duration(1) * time.Hour) + (time.Duration(2) * time.Minute) + (time.Duration(3) * time.Second),
		Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", int(7*int64(time.Hour/time.Second)))),
	}

	err := (&scanner{}).Scan(&f)
	if err != nil {
		t.Error(err)
	}

	verify(f, t)
}

func TestDefault(t *testing.T) {
	os.Clearenv()

	const s = `owls are not what do you think about them`

	var f struct {
		String string `default:"owls are not what do you think about them"`
	}
	err := Env("app_").Scan(&f)
	if err != nil {
		t.Error(err)
	}
	if f.String != s {
		t.Errorf("f.String = '%v', want '%s'", f.String, s)
	}
}

func TestEnvScan(t *testing.T) {
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
	err := Env("app_").Scan(&f)
	if err != nil {
		t.Error(err)
	}

	verify(f, t)
}

func testErrEnvScan(k string, t *testing.T) {
	os.Clearenv()
	os.Setenv(k, "Err")

	var f Fields
	err := Env("app_").Scan(&f)
	if err == nil {
		t.Errorf("err=nil for errorneous %s, want not nil", k)
	}
	// extend coverage a little bit more
	if err.Error() == "" {
		t.Error("err.Error()=\"\", want non empty string")
	}
}

func TestEnvScan_UintErr(t *testing.T) {
	testErrEnvScan("APP_UINT", t)
}

func TestEnvScan_IntErr(t *testing.T) {
	testErrEnvScan("APP_INT", t)
}

func TestEnvScan_FloatErr(t *testing.T) {
	testErrEnvScan("APP_FLOAT", t)
}

func TestEnvScan_BoolErr(t *testing.T) {
	testErrEnvScan("APP_BOOL", t)
}

func TestEnvScan_DurationErr(t *testing.T) {
	testErrEnvScan("APP_DURATION", t)
}

func TestEnvScan_DateErr(t *testing.T) {
	testErrEnvScan("APP_DATE", t)
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

func CreateFile(content string) (*os.File, error) {
	file, err := ioutil.TempFile(os.TempDir(), "appconfig_test")
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(file, content); err != nil {
		return nil, err
	}
	return file, err
}

func TestFileScan(t *testing.T) {
	file, err := CreateFile(stream)
	defer func() {
		if file != nil {
			os.Remove(file.Name())
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	var f Fields
	if err := File(file.Name()).Scan(&f); err != nil {
		t.Fatal(err)
	}

	verify(f, t)
}

func TestFileScan_NoFile(t *testing.T) {
	var f Fields
	if err := File(":$:").Scan(&f); err != nil {
		t.Errorf("err = '%s', want nil", err)
	}
}

func TestReaderScan(t *testing.T) {
	var f Fields
	if err := Reader(bytes.NewBufferString(stream)).Scan(&f); err != nil {
		t.Error(err)
	}

	verify(f, t)
}

func TestReaderScan_Error(t *testing.T) {
	var f Fields
	if Reader(iotest.TimeoutReader(bytes.NewBufferString(stream))).Scan(&f) == nil {
		t.Errorf("err = nil, want not nil")
	}
}

func TestEnvReaderScan(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_STRING", "123.4")
	reader := bytes.NewBufferString("int=-123")

	var f Fields
	if err := Env("app_").Reader(reader).Scan(&f); err != nil {
		t.Error(err)
	}

	if f.String != "123.4" {
		t.Errorf("f.String = '%v', want '123.4'", f.String)
	}
	if f.Int != -123 {
		t.Errorf("f.Int = %v, want -123", f.Int)
	}
}

func TestReaderEnvScan(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_STRING", "123.4")
	reader := bytes.NewBufferString("int=-123")

	var f Fields
	if err := Reader(reader).Env("app_").Scan(&f); err != nil {
		t.Error(err)
	}

	if f.String != "123.4" {
		t.Errorf("f.String = '%v', want '123.4'", f.String)
	}
	if f.Int != -123 {
		t.Errorf("f.Int = %v, want -123", f.Int)
	}
}
