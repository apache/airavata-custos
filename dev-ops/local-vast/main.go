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

// Command local-vast serves the part of the VAST VMS REST API the
// VAST-Provisioner calls, on a test cluster's file server. Folders become
// real directories whose owners resolve through NSS, as VMS resolves them
// through its directory provider; views become NFS exports and quotas are only recorded.
package main

import (
	"cmp"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	user, password, token := os.Getenv("VAST_USER"), os.Getenv("VAST_PASSWORD"), rand.Text()
	exportTo, options := os.Getenv("EXPORT_TO"), cmp.Or(os.Getenv("EXPORT_OPTIONS"), "rw,sync,no_subtree_check,root_squash")
	if user == "" || password == "" || exportTo == "" {
		log.Fatal("local-vast: VAST_USER, VAST_PASSWORD and EXPORT_TO are required")
	}
	authed := func(h func(r *http.Request, body []byte) (any, error)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body json.RawMessage
			_ = json.NewDecoder(r.Body).Decode(&body)
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "token not valid", http.StatusForbidden)
			} else if out, err := h(r, body); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			} else if out != nil {
				_ = json.NewEncoder(w).Encode(out)
			}
		}
	}
	var mu sync.Mutex
	quotas, lastQuotaID := []map[string]any{}, 0
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
	mux.HandleFunc("POST /api/folders/create_folder/", authed(func(_ *http.Request, body []byte) (any, error) { return nil, createFolder(body) }))
	mux.HandleFunc("POST /api/folders/stat_path/", authed(func(_ *http.Request, body []byte) (any, error) {
		var b struct{ Path string }
		_ = json.Unmarshal(body, &b)
		fi, err := os.Stat(b.Path)
		if err != nil {
			return nil, err
		}
		st := fi.Sys().(*syscall.Stat_t)
		return map[string]any{"owning_uid": st.Uid, "owning_gid": st.Gid, "is_directory": fi.IsDir()}, nil
	}))
	// VAST's trash keeps a deleted folder until it empties itself; here it moves into .trash beside it.
	mux.HandleFunc("DELETE /api/folders/delete_folder/", authed(func(_ *http.Request, body []byte) (any, error) {
		var b struct{ Path string }
		_ = json.Unmarshal(body, &b)
		trash := filepath.Join(filepath.Dir(b.Path), ".trash")
		if err := os.MkdirAll(trash, 0o700); err != nil {
			return nil, err
		}
		return nil, os.Rename(b.Path, filepath.Join(trash, filepath.Base(b.Path)+"-"+strconv.FormatInt(time.Now().Unix(), 10)))
	}))
	mux.HandleFunc("POST /api/quotas/", authed(func(_ *http.Request, body []byte) (any, error) {
		var q map[string]any
		_ = json.Unmarshal(body, &q)
		mu.Lock()
		defer mu.Unlock()
		lastQuotaID++
		q["id"], q["tenant_id"] = lastQuotaID, cmp.Or(q["tenant_id"], any(1))
		quotas = append(quotas, q)
		return q, nil
	}))
	mux.HandleFunc("GET /api/quotas/", authed(func(r *http.Request, _ []byte) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		found := slices.DeleteFunc(slices.Clone(quotas), func(q map[string]any) bool { return !matches(r, q) })
		for i, q := range found {
			found[i] = maps.Clone(q)
			found[i]["used_capacity"] = used(fmt.Sprint(q["path"]))
		}
		return found, nil
	}))
	mux.HandleFunc("DELETE /api/quotas/{id}/", authed(func(r *http.Request, _ []byte) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		quotas = slices.DeleteFunc(quotas, func(q map[string]any) bool { return fmt.Sprint(q["id"]) == r.PathValue("id") })
		return nil, nil
	}))
	// Views are NFS exports to EXPORT_TO, identified by their directory's inode so nothing is stored.
	mux.HandleFunc("POST /api/views/", authed(func(_ *http.Request, body []byte) (any, error) {
		var b struct{ Path string }
		_ = json.Unmarshal(body, &b)
		_, err := exportfs("-o", options, exportTo+":"+b.Path)
		return nil, err
	}))
	mux.HandleFunc("GET /api/views/", authed(func(r *http.Request, _ []byte) (any, error) {
		vs, err := views()
		return slices.DeleteFunc(vs, func(v map[string]any) bool { return !matches(r, v) }), err
	}))
	mux.HandleFunc("DELETE /api/views/{id}/", authed(func(r *http.Request, _ []byte) (any, error) {
		vs, err := views()
		i := slices.IndexFunc(vs, func(v map[string]any) bool { return fmt.Sprint(v["id"]) == r.PathValue("id") })
		if err != nil || i < 0 {
			return nil, cmp.Or(err, fmt.Errorf("no view %s", r.PathValue("id")))
		}
		_, err = exportfs("-u", exportTo+":"+vs[i]["path"].(string))
		return nil, err
	}))
	// Every policy name exists here, so the connector's startup check passes.
	mux.HandleFunc("GET /api/viewpolicies/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "token not valid", http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": r.URL.Query().Get("name")}})
	})
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

func views() ([]map[string]any, error) {
	out, err := exportfs("-s")
	vs := []map[string]any{}
	for line := range strings.Lines(out) {
		path, _, _ := strings.Cut(strings.TrimSpace(line), " ")
		if fi, err := os.Stat(path); err == nil {
			vs = append(vs, map[string]any{"id": fi.Sys().(*syscall.Stat_t).Ino, "path": path, "tenant_id": 1, "protocols": []string{"NFS"}, "policy": "local"})
		}
	}
	return vs, err
}

func used(path string) int64 {
	var n int64
	_ = filepath.Walk(path, func(_ string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() {
			n += fi.Size()
		}
		return nil
	})
	return n
}

func matches(r *http.Request, e map[string]any) bool {
	for k, v := range r.URL.Query() {
		if fmt.Sprint(e[k]) != v[0] {
			return false
		}
	}
	return true
}

func exportfs(args ...string) (string, error) {
	out, err := exec.Command("exportfs", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exportfs: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}
