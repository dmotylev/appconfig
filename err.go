// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig

import (
	"reflect"
)

// Err wraps error in Source interface
func Err(e error) Source {
	return &errSrc{e}
}

type errSrc struct {
	e error
}

func (s *errSrc) Err() error {
	return s.e
}

func (s *errSrc) Lookup(f reflect.StructField) (string, bool) {
	return "", false
}
