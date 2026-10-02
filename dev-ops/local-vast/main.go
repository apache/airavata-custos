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

// Command local-vast serves the part of the VAST VMS REST API the VAST
// Storage-Provisioner calls, on a test cluster's file server. Folders become
// real directories whose owners resolve through NSS, as VMS resolves them
// through its directory provider; quotas are only logged.
package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	user, password, token := os.Getenv("VAST_USER"), os.Getenv("VAST_PASSWORD"), rand.Text()
	if user == "" || password == "" {
		log.Fatal("local-vast: VAST_USER and VAST_PASSWORD are required")
	}
	authed := func(h func(body []byte) error) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body json.RawMessage
			_ = json.NewDecoder(r.Body).Decode(&body)
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "token not valid", http.StatusForbidden)
			} else if err := h(body); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/token/", func(w http.ResponseWriter, r *http.Request) {
		var b struct{ Username, Password string }
		_ = json.NewDecoder(r.Body).Decode(&b)
		if b.Username != user || b.Password != password {
			http.Error(w, "authentication failed", http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"access": token})
	})
	mux.HandleFunc("POST /api/folders/create_folder/", authed(createFolder))
	mux.HandleFunc("POST /api/folders/stat_path/", authed(func(body []byte) error {
		var b struct{ Path string }
		_ = json.Unmarshal(body, &b)
		_, err := os.Stat(b.Path)
		return err
	}))
	mux.HandleFunc("POST /api/quotas/", authed(func(body []byte) error {
		log.Printf("quota %s", body)
		return nil
	}))
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN"), mux))
}

func createFolder(body []byte) error {
	var b struct {
		Path, User, Group string
		Mode              string `json:"create_dir_mode"`
	}
	_ = json.Unmarshal(body, &b)
	uid, err := getent("passwd", b.User)
	if err != nil {
		return errors.New("Couldn't find user " + b.User)
	}
	gid, err := getent("group", b.Group)
	if err != nil {
		return errors.New("Couldn't find group " + b.Group)
	}
	mode, _ := strconv.ParseUint(b.Mode, 8, 32)
	// Chmod after Mkdir, since Mkdir is subject to the umask and drops setgid.
	if err := os.Mkdir(b.Path, 0o700); err != nil {
		return err
	}
	return errors.Join(os.Chown(b.Path, uid, gid), syscall.Chmod(b.Path, uint32(mode)))
}

func getent(db, name string) (int, error) {
	out, err := exec.Command("getent", db, name).Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.Split(string(out), ":")[2])
}
