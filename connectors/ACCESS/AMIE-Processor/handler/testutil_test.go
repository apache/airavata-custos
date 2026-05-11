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

package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// testdataDir returns the absolute path to the module-level testdata directory.
// It uses runtime.Caller so the path is correct regardless of where tests run.
func testdataDir() string {
	// handler/ is one level below the module root; go up one directory.
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "testdata")
}

// loadTestData reads a JSON fixture file from testdata/<subpath> and unmarshals
// it into a map[string]any.  The test fails immediately if the file cannot be
// read or parsed.
func loadTestData(t *testing.T, subpath string) map[string]any {
	t.Helper()
	path := filepath.Join(testdataDir(), subpath)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "reading fixture %s", path)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m), "parsing fixture %s", path)
	return m
}
