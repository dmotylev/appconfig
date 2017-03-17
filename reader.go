// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var keyValueRE = regexp.MustCompile("^([^=]+)=(.*)$")

// FromReader utilizes reader as a source for lookup.
// For any line matched to 'key=value' pattern it creates an entry in the internal map.
// Then map is used as the source for values.
func FromReader(reader io.Reader) Source {
	m := make(map[string]string)

	s := bufio.NewScanner(bufio.NewReader(reader))
	for s.Scan() {
		l := s.Text()
		p := keyValueRE.FindStringSubmatch(l)
		if p == nil {
			continue
		}
		m[strings.ToLower(p[1])] = p[2]
	}
	if err := s.Err(); err != nil {
		return Err(err)
	}
	return FromMap(m)
}
