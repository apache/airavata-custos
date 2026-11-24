// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package krl builds OpenSSH key revocation lists (PROTOCOL.krl).
package krl

import (
	"encoding/binary"
	"time"
)

// Build returns a KRL that revokes the given certificate serials, whichever CA signed them.
// Serials never repeat across CA rotations, so one list covers every CA.
func Build(version uint64, generated time.Time, serials []int64) []byte {
	b := []byte("SSHKRL\n\x00")
	b = binary.BigEndian.AppendUint32(b, 1)
	b = binary.BigEndian.AppendUint64(b, version)
	b = binary.BigEndian.AppendUint64(b, uint64(generated.Unix()))
	b = binary.BigEndian.AppendUint64(b, 0) // flags
	b = appendString(b, nil)                // reserved
	b = appendString(b, nil)                // comment
	if len(serials) == 0 {
		return b
	}

	var list []byte
	for _, serial := range serials {
		list = binary.BigEndian.AppendUint64(list, uint64(serial))
	}
	var section []byte
	section = appendString(section, nil) // CA key, empty means any CA
	section = appendString(section, nil) // reserved
	section = append(section, 0x20)      // serial list
	section = appendString(section, list)

	b = append(b, 1) // certificates section
	return appendString(b, section)
}

func appendString(b, s []byte) []byte {
	b = binary.BigEndian.AppendUint32(b, uint32(len(s)))
	return append(b, s...)
}
