// Copyright (c) 2013 Dmitry Motylev. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package appconfig

import "os"

// FromFile uses Reader if the file exists, empty map is the source otherwise.
func FromFile(s string) Source {
	f, err := os.Open(s)
	if err != nil {
		return FromMap(nil)
	}
	defer f.Close()
	return FromReader(f)
}
