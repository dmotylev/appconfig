// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig

import (
	"os"
	"reflect"
	"strings"
)

// FromEnv utilizes environment variables for lookup
func FromEnv(p string) Source {
	return &envSrc{p}
}

type envSrc struct {
	prefix string
}

func (s *envSrc) Err() error {
	return nil
}

func (s *envSrc) Lookup(f reflect.StructField) (string, bool) {
	return os.LookupEnv(strings.ToUpper(s.prefix + f.Name))
}
