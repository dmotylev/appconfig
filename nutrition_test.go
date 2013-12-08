package nutrition

import (
	"os"
	"testing"
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

func TestHarvestPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic on non-struct type, want panic")
		}
	}()
	var i int
	(&harvester{}).Harvest(i)
}

func TestHarvest(t *testing.T) {
	f := Fields{
		String:   "123.4",
		Uint:     123,
		Int:      -123,
		Float:    123.4,
		Bool:     true,
		Duration: (time.Duration(1) * time.Hour) + (time.Duration(2) * time.Minute) + (time.Duration(3) * time.Second),
		Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", int(7*int64(time.Hour/time.Second)))),
	}

	err := (&harvester{}).Harvest(&f)
	if err != nil {
		t.Error(err)
	}

	verify(f, t)
}

func TestEnvHarvest(t *testing.T) {
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
	err := Env("app_").Harvest(&f)
	if err != nil {
		t.Error(err)
	}

	verify(f, t)
}

func testErrEnvHarvest(k string, t *testing.T) {
	os.Clearenv()
	os.Setenv(k, "Err")

	var f Fields
	err := Env("app_").Harvest(&f)
	if err == nil {
		t.Errorf("err=nil for errorneous %s, want not nil", k)
	}
	// extend coverage a little bit more
	if err.Error() == "" {
		t.Error("err.Error()=\"\", want non empty string")
	}
}

func TestEnvHarvest_UintErr(t *testing.T) {
	testErrEnvHarvest("APP_UINT", t)
}

func TestEnvHarvest_IntErr(t *testing.T) {
	testErrEnvHarvest("APP_INT", t)
}

func TestEnvHarvest_FloatErr(t *testing.T) {
	testErrEnvHarvest("APP_FLOAT", t)
}

func TestEnvHarvest_BoolErr(t *testing.T) {
	testErrEnvHarvest("APP_BOOL", t)
}

func TestEnvHarvest_DurationErr(t *testing.T) {
	testErrEnvHarvest("APP_DURATION", t)
}

func TestEnvHarvest_DateErr(t *testing.T) {
	testErrEnvHarvest("APP_DATE", t)
}
