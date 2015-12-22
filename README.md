# Nutrition [![Build Status](https://travis-ci.org/dmotylev/nutrition.png?branch=master)](https://travis-ci.org/dmotylev/nutrition) [![Coverage Status](https://coveralls.io/repos/dmotylev/nutrition/badge.png)](https://coveralls.io/r/dmotylev/nutrition) [![GoDoc](https://godoc.org/github.com/dmotylev/appconfig?status.svg)](https://godoc.org/github.com/dmotylev/appconfig)

Package nutrition provides decoding of different sources based on user defined struct.
Source is the stream of lines in 'key=value' form. Environment, file and raw io.Reader sources are supported.
Package can decode types boolean, numeric, and string types as well as time.Duration and time.Time. The later could be customized with formats.

# Documentation

The Nutrition API reference is available on [GoDoc](http://godoc.org/github.com/dmotylev/nutrition).

# Installation

Install Nutrition using the `go get` command:

	go get -u github.com/dmotylev/nutrition

# Example

Feed struct from environment only:

```go
// Try:
//
// APP_TIMEOUT=1h2m3s APP_DAY=2013-12-13 APP_NUMWORKERS=5 APP_WORKERNAME=Hulk go run p.go
//
package main

import (
	"fmt"
	"time"

	"github.com/dmotylev/nutrition"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := nutrition.Env("APP_").Feed(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```

Feed struct from stdin only:

```go
// Try:
//
// echo -e "timeout=1h2m3s\nday=2013-12-13\nnumWorkers=5\nworkerName=Hulk" | go run p.go
//
package main

import (
	"fmt"
	"time"
	"os"

	"github.com/dmotylev/nutrition"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := nutrition.Reader(os.Stdin).Feed(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```

Feed struct from file only:

```go
// Try:
//
// echo -e "timeout=1h2m3s\nday=2013-12-13\nnumWorkers=5\nworkerName=Hulk" > env && go run p.go
//
package main

import (
	"fmt"
	"time"

	"github.com/dmotylev/nutrition"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := nutrition.Reader("env").Feed(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```

All together:

```go
// Try:
//
// echo -e "timeout=1h2m3s\nday=2013-12-13\nnumWorkers=5\nworkerName=Hulk" > env && echo -e "numWorkers=10000\nworkerName=Fido" | APP_WORKERNAME=Pinnoccio go run p.go
//
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dmotylev/nutrition"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	// You can change the order. First source with value will be used
	err := nutrition.Env("APP_").Reader(os.Stdin).File("env").Feed(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```
# License

The package available under [LICENSE](https://github.com/dmotylev/nutrition/blob/master/LICENSE).
