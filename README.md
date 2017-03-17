# AppConfig [![Build Status](https://travis-ci.org/dmotylev/appconfig.png?branch=master)](https://travis-ci.org/dmotylev/appconfig) [![Coverage Status](https://coveralls.io/repos/dmotylev/appconfig/badge.png)](https://coveralls.io/r/dmotylev/appconfig) [![Go Report Card](https://goreportcard.com/badge/github.com/dmotylev/appconfig)](https://goreportcard.com/report/github.com/dmotylev/appconfig) [![GoDoc](https://godoc.org/github.com/dmotylev/appconfig?status.svg)](https://godoc.org/github.com/dmotylev/appconfig) 

```Go
import "github.com/dmotylev/appconfig"
```

## Documentation

See [GoDoc](http://godoc.org/github.com/dmotylev/appconfig).

## Usage

Set environment variables:
```Bash
APP_TIMEOUT=1h2m3s
APP_WORKERNAME=Monkey
```

Write other values to the file:
```Bash
cat > local.conf << EOF
APP_TIMEOUT=2h2m3s
APP_NUMWORKERS=10
APP_WORKERNAME=Robot
EOF
```

Write code:

Populate variable from both sources:

```Go
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

	err := appconfig.Load(&conf, appconfig.FromEnv("APP_"), appconfig.FromFile("local.conf"))

	fmt.Printf("err=%v\n", err)
	fmt.Printf("timeout=%s\n", conf.Timeout)
	fmt.Printf("day=%s\n", conf.Day.Format(time.UnixDate))
	fmt.Printf("worker=%s\n", conf.WorkerName)
	fmt.Printf("workers=%d\n", conf.NumWorkers)
}
```

Results:
```Bash
err=<nil>
timeout=1h2m3s
day=Fri Dec 13 00:00:00 UTC 2013
worker=Monkey
workers=10
```

Get some inspiration from tests.

# License

The package available under [LICENSE](https://github.com/dmotylev/appconfig/blob/master/LICENSE).
