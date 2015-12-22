# AppConfig [![Build Status](https://travis-ci.org/dmotylev/appconfig.png?branch=master)](https://travis-ci.org/dmotylev/appconfig) [![Coverage Status](https://coveralls.io/repos/dmotylev/appconfig/badge.png)](https://coveralls.io/r/dmotylev/appconfig) [![GoDoc](https://godoc.org/github.com/dmotylev/appconfig?status.svg)](https://godoc.org/github.com/dmotylev/appconfig)

Package appconfig provides initialization for user defined struct from configured sources.
Source is the stream of lines in 'key=value' form. Environment, file and raw 
io.Reader sources are supported. The package can decode boolean, numeric, string,
time.Duration and time.Time. The later could be customized with formats.

# Documentation

The AppConfig API reference is available on [GoDoc](http://godoc.org/github.com/dmotylev/appconfig).

# Installation

Install AppConfig using the `go get` command:

	go get -u github.com/dmotylev/appconfig

# Example

Scan struct from environment only:

```go
// Try:
//
// APP_TIMEOUT=1h2m3s APP_DAY=2013-12-13 APP_NUMWORKERS=5 APP_WORKERNAME=Hulk go run p.go
//
package main

import (
	"fmt"
	"time"

	"github.com/dmotylev/appconfig"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := appconfig.Env("APP_").Scan(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```

Scan struct from stdin only:

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

	"github.com/dmotylev/appconfig"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := appconfig.Reader(os.Stdin).Scan(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```

Scan struct from file only:

```go
// Try:
//
// echo -e "timeout=1h2m3s\nday=2013-12-13\nnumWorkers=5\nworkerName=Hulk" > env && go run p.go
//
package main

import (
	"fmt"
	"time"

	"github.com/dmotylev/appconfig"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	err := appconfig.Reader("env").Scan(&conf)

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

	"github.com/dmotylev/appconfig"
)

func main() {
	var conf struct {
		Timeout    time.Duration
		Day        time.Time `time,format:"2006-01-02"`
		WorkerName string
		NumWorkers int
	}

	// You can change the order. First source with value will be used
	err := appconfig.Env("APP_").Reader(os.Stdin).File("env").Scan(&conf)

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```
# License

The package available under [LICENSE](https://github.com/dmotylev/appconfig/blob/master/LICENSE).
