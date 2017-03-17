// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig

import (
	"reflect"
	"strings"
)

// FromMap looks values in the map
func FromMap(m map[string]string) Source {
	return &mapSrc{m}
}

type mapSrc struct {
	m map[string]string
}

func (s *mapSrc) Err() error {
	return nil
}

func (s *mapSrc) Lookup(f reflect.StructField) (string, bool) {
	v, found := s.m[strings.ToLower(f.Name)]
	return v, found
}
