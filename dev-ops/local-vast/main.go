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
// through its directory provider. Shell commands do the work: views become NFS
// exports to EXPORT_TO, identified by their directory's inode so nothing is
// stored, and quotas, which a plain file system cannot enforce, are only recorded.
package main

import (
	"cmp"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type request struct {
	Path, User, Group string
	Mode              string `json:"create_dir_mode"`
}

func decode(body []byte) (r request) {
	_ = json.Unmarshal(body, &r)
	return r
}

type view struct {
	ID        int      `json:"id"`
	Path      string   `json:"path"`
	TenantID  int      `json:"tenant_id"`
	Protocols []string `json:"protocols"`
	Policy    string   `json:"policy"`
}

func main() {
	user, password, token := os.Getenv("VAST_USER"), os.Getenv("VAST_PASSWORD"), rand.Text()
	exportTo, options := os.Getenv("EXPORT_TO"), cmp.Or(os.Getenv("EXPORT_OPTIONS"), "rw,sync,no_subtree_check,root_squash")
	if user == "" || password == "" || exportTo == "" {
		log.Fatal("local-vast: VAST_USER, VAST_PASSWORD and EXPORT_TO are required")
	}
	var mu sync.Mutex
	quotas, lastQuotaID := []map[string]any{}, 0
	authed := func(h func(r *http.Request, body []byte) (any, error)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body json.RawMessage
			_ = json.NewDecoder(r.Body).Decode(&body)
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "token not valid", http.StatusForbidden)
			} else if out, err := h(r, body); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
			} else if out != nil {
				_ = json.NewEncoder(w).Encode(out)
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
	mux.HandleFunc("POST /api/folders/stat_path/", authed(func(_ *http.Request, body []byte) (any, error) {
		var out []byte
		if err := run(&out, `{ test -e "$1" || { echo "Template directory path wasn't found" >&2; exit 1; }; } && stat -c '%u %g %F' "$1"`, decode(body).Path); err != nil {
			return nil, err
		}
		f := strings.SplitN(strings.TrimSpace(string(out)), " ", 3)
		return map[string]any{"owning_uid": f[0], "owning_gid": f[1], "is_directory": f[2] == "directory"}, nil
	}))
	mux.HandleFunc("POST /api/quotas/", authed(func(_ *http.Request, body []byte) (any, error) {
		var q map[string]any
		_ = json.Unmarshal(body, &q)
		mu.Lock()
		defer mu.Unlock()
		if slices.ContainsFunc(quotas, func(e map[string]any) bool { return e["name"] == q["name"] }) {
			return nil, fmt.Errorf("quota %v exists", q["name"])
		}
		lastQuotaID++
		q["id"] = lastQuotaID
		if q["tenant_id"] == nil {
			q["tenant_id"] = 1
		}
		quotas = append(quotas, q)
		return q, nil
	}))
	mux.HandleFunc("DELETE /api/quotas/{id}/", authed(func(r *http.Request, _ []byte) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		quotas = slices.DeleteFunc(quotas, func(q map[string]any) bool { return fmt.Sprint(q["id"]) == r.PathValue("id") })
		return nil, nil
	}))
	mux.HandleFunc("GET /api/quotas/", authed(func(r *http.Request, _ []byte) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		return slices.DeleteFunc(slices.Clone(quotas), func(q map[string]any) bool {
			return !matches(r, "name", q["name"]) || !matches(r, "path", q["path"]) || !matches(r, "tenant_id", q["tenant_id"])
		}), nil
	}))
	mux.HandleFunc("POST /api/views/", authed(func(_ *http.Request, body []byte) (any, error) {
		return nil, run(nil, `exportfs -o "$1" "$2:$3"`, options, exportTo, decode(body).Path)
	}))
	mux.HandleFunc("GET /api/views/", authed(func(r *http.Request, _ []byte) (any, error) {
		vs, err := views()
		return slices.DeleteFunc(vs, func(v view) bool { return !matches(r, "path", v.Path) || !matches(r, "tenant_id", v.TenantID) }), err
	}))
	mux.HandleFunc("DELETE /api/views/{id}/", authed(func(r *http.Request, _ []byte) (any, error) {
		vs, err := views()
		i := slices.IndexFunc(vs, func(v view) bool { return strconv.Itoa(v.ID) == r.PathValue("id") })
		if err != nil || i < 0 {
			return nil, cmp.Or(err, fmt.Errorf("no view %s", r.PathValue("id")))
		}
		return nil, run(nil, `exportfs -u "$1:$2"`, exportTo, vs[i].Path)
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

func createFolder(_ *http.Request, body []byte) (any, error) {
	b := decode(body)
	// Five digits make chmod set setgid exactly instead of keeping the one inherited from the parent.
	return nil, run(nil, `{ test ! -e "$1" || { echo "$1 exists" >&2; exit 1; }; } && id -u "$2" >/dev/null &&
		getent group "$3" >/dev/null && mkdir "$1" && chown "$2:$3" "$1" && chmod "$(printf %05d "$4")" "$1"`,
		b.Path, b.User, b.Group, b.Mode)
}

// views lists the NFS exports as views.
func views() ([]view, error) {
	var out []byte
	if err := run(&out, `exportfs -s | while read -r p _; do echo "$(stat -c %i "$p") $p"; done`); err != nil {
		return nil, err
	}
	vs := []view{}
	for line := range strings.Lines(string(out)) {
		ino, path, _ := strings.Cut(strings.TrimSpace(line), " ")
		id, _ := strconv.Atoi(ino)
		vs = append(vs, view{ID: id, Path: path, TenantID: 1, Protocols: []string{"NFS"}, Policy: "local"})
	}
	return vs, nil
}

// matches reports whether the query parameter key is unset or equals v.
func matches(r *http.Request, key string, v any) bool {
	q := r.URL.Query().Get(key)
	return q == "" || q == fmt.Sprint(v)
}

// run runs script under sh with args as $1..., putting its standard output in out when given.
func run(out *[]byte, script string, args ...string) error {
	cmd := exec.Command("sh", append([]string{"-c", script, "sh"}, args...)...)
	stdout, err := cmd.Output()
	if out != nil {
		*out = stdout
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(ee.Stderr)), err)
	}
	return err
}
