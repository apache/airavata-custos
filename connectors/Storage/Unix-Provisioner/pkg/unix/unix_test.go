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

package unix

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

func TestPlans(t *testing.T) {
	ctx, group, now := context.Background(), "proj-a", time.Now()
	alloc := models.ComputeAllocation{ID: "a1", ComputeClusterID: "c1", Status: models.ACTIVE, PosixGroup: &group}
	member := func(name string, status models.AllocationStatus, at *time.Time) store.MembershipWithUser {
		return store.MembershipWithUser{ComputeAllocationMembership: models.ComputeAllocationMembership{MembershipStatus: status}, LocalUsername: name, ProvisionedAt: at}
	}
	p := &provisioner{clusters: map[string]cluster{"c1": {ID: "c1", Targets: []dir{{Path: "/home/{user}"}, {Path: "/project/{allocation}"}, {Path: "/project/{allocation}/{user}"}}}},
		core: &service.CoreServiceMock{
			ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) {
				return []models.ComputeAllocationMembership{{ComputeAllocationID: "a1"}}, nil
			},
			GetComputeAllocationFunc: func(context.Context, string) (*models.ComputeAllocation, error) { return &alloc, nil },
			ListMembersForAllocationFunc: func(context.Context, string) ([]store.MembershipWithUser, error) {
				return []store.MembershipWithUser{member("jdoe", models.ACTIVE, &now), member("unprov", models.ACTIVE, nil), member("gone", models.INACTIVE, &now)}, nil
			},
		}}
	plan := func(_ cluster, dirs []dir, err error) string {
		var got []string
		for _, d := range dirs {
			got = append(got, d.Path)
		}
		return fmt.Sprint(got, err)
	}
	for _, c := range [][2]string{
		{plan(p.account(ctx, models.ComputeClusterUser{ComputeClusterID: "c1", LocalUsername: "jdoe"})), "[] <nil>"},
		{plan(p.account(ctx, models.ComputeClusterUser{ComputeClusterID: "c2", LocalUsername: "jdoe", ProvisionedAt: &now})), "[] <nil>"},
		{plan(p.account(ctx, models.ComputeClusterUser{ComputeClusterID: "c1", LocalUsername: "jdoe", ProvisionedAt: &now})), "[/home/jdoe /project/proj-a /project/proj-a/jdoe] <nil>"},
		{plan(p.allocation(ctx, "a1")), "[/project/proj-a /project/proj-a/jdoe] <nil>"},
	} {
		if c[0] != c[1] {
			t.Errorf("got %s; want %s", c[0], c[1])
		}
	}
}

func TestCommand(t *testing.T) {
	c := cluster{ID: "c1"}
	c.SSH.User, c.SSH.Hostname, c.SSH.Identity = "root", "admin", "/k"
	tmp, uid, gid := t.TempDir(), strconv.Itoa(os.Getuid()), strconv.Itoa(os.Getgid())
	d := func(path, owner string) dir { return dir{tmp + path, owner, gid, "2750"} }
	args, script, err := command(c, []dir{d("/a", uid), d("/a/b", "nosuchuser"), d("/x/y", uid), d("/a;reboot", uid)})
	sh := exec.Command("sh")
	sh.Stdin = strings.NewReader(script)
	out, _ := sh.Output()
	a, _ := os.Stat(tmp + "/a")
	_, statErr := os.Stat(tmp + "/a/b")
	if fmt.Sprint(args) != "[-o BatchMode=yes -i /k root@admin sh]" || string(out) != tmp+"/a\n" || a.Mode()&os.ModeSetgid == 0 || !os.IsNotExist(statErr) || err == nil {
		t.Fatalf("args %v, printed %q, a %v, a/b %v, refusal %v", args, out, a.Mode(), statErr, err)
	}
}
