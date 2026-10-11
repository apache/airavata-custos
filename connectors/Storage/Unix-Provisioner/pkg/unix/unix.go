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

// Package unix is the Unix-Provisioner: it creates each configured cluster's directories over SSH on an admin host
// that mounts the storage, an alternative to the VAST-Provisioner for storage without a management API. The host's NSS
// resolves owners, so a directory whose owner it does not know yet is removed again and the bus retries it. Each
// trigger plans one cluster's directories in config order, parents first, and creates them in one SSH session.
package unix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "unix-storage-provisioner"

type cluster struct {
	ID      string
	SSH     struct{ User, Hostname, Identity string }
	Targets []dir
}

// dir is a directory per account, per allocation, or per member of one. Path, Owner and Group are templates over
// {user} and {allocation}; the ones it uses decide which.
type dir struct{ Path, Owner, Group, Mode string }

type provisioner struct {
	core     service.CoreService
	clusters map[string]cluster
}

func LoadConnector(ctx context.Context, _ *sqlx.DB, bus *events.Bus, core *service.Service, wg *sync.WaitGroup, _ *identity.Router, cfg *config.ConnectorConfig) error {
	var list []cluster
	b, _ := json.Marshal(cfg.Config["clusters"])
	_ = json.Unmarshal(b, &list)
	clusters := map[string]cluster{}
	for _, c := range list {
		if c.ID == "" || c.SSH.User == "" || c.SSH.Hostname == "" || len(c.Targets) == 0 {
			slog.Info("unix provisioner: required config not set; skipping", "cluster_id", c.ID)
			continue
		}
		clusters[c.ID] = c
		slog.Info("unix provisioner: cluster registered", "host", c.SSH.Hostname, "cluster_id", c.ID)
	}
	if len(clusters) == 0 {
		return nil
	}
	p := &provisioner{core, clusters}
	byAllocation := handle(func(ctx context.Context, a models.ComputeAllocation) (cluster, []dir, error) {
		return p.allocation(ctx, a.ID)
	})
	byMembership := handle(func(ctx context.Context, m models.ComputeAllocationMembership) (cluster, []dir, error) {
		return p.allocation(ctx, m.ComputeAllocationID)
	})
	bus.SubscribeComputeClusterUserUpdated(Type, handle(p.account))
	bus.SubscribeComputeAllocationCreated(Type, byAllocation)
	bus.SubscribeComputeAllocationUpdated(Type, byAllocation)
	bus.SubscribeComputeAllocationMembershipCreated(Type, byMembership)
	bus.SubscribeComputeAllocationMembershipUpdated(Type, byMembership)
	wg.Go(func() {
		for {
			for id, c := range clusters {
				ctx, cancel := context.WithTimeout(ctx, time.Hour)
				if err := handle(p.reconcile)(ctx, c); err != nil {
					slog.Error("unix provisioner: reconcile", "cluster_id", id, "error", err)
				}
				cancel()
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(24 * time.Hour):
			}
		}
	})
	return nil
}

func handle[T any](plan func(context.Context, T) (cluster, []dir, error)) func(context.Context, T) error {
	return func(ctx context.Context, t T) error {
		c, dirs, err := plan(ctx, t)
		return errors.Join(err, ensure(ctx, c, dirs))
	}
}

// account needs no approval check: provisioned_at is set only after approval.
func (p *provisioner) account(ctx context.Context, cu models.ComputeClusterUser) (cluster, []dir, error) {
	c, ok := p.clusters[cu.ComputeClusterID]
	if !ok || cu.ProvisionedAt == nil {
		return c, nil, nil
	}
	dirs := dirsFor(c, cu.LocalUsername, "")
	memberships, err := p.core.ListAllocationsForUser(ctx, cu.UserID)
	for _, m := range memberships {
		ac, ad, e := p.allocation(ctx, m.ComputeAllocationID)
		if ac.ID == c.ID {
			dirs = append(dirs, ad...)
		}
		err = errors.Join(err, e)
	}
	return c, dirs, err
}

// allocation plans an active allocation's directories, then each active, provisioned member's.
func (p *provisioner) allocation(ctx context.Context, id string) (cluster, []dir, error) {
	a, err := p.core.GetComputeAllocation(ctx, id)
	if errors.Is(err, service.ErrNotFound) {
		return cluster{}, nil, nil
	}
	if err != nil {
		return cluster{}, nil, err
	}
	c, ok := p.clusters[a.ComputeClusterID]
	if !ok || a.Status != models.ACTIVE || a.PosixGroup == nil {
		return c, nil, nil
	}
	dirs := dirsFor(c, "", *a.PosixGroup)
	members, err := p.core.ListMembersForAllocation(ctx, a.ID)
	for _, m := range members {
		if m.MembershipStatus == models.ACTIVE && m.ProvisionedAt != nil {
			dirs = append(dirs, dirsFor(c, m.LocalUsername, *a.PosixGroup)...)
		}
	}
	return c, dirs, err
}

func (p *provisioner) reconcile(ctx context.Context, c cluster) (cluster, []dir, error) {
	var dirs []dir
	users, err := p.core.ListComputeClusterUsersByCluster(ctx, c.ID)
	for _, cu := range users {
		if cu.ProvisionedAt != nil {
			dirs = append(dirs, dirsFor(c, cu.LocalUsername, "")...)
		}
	}
	allocs, e := p.core.ListComputeAllocationsByCluster(ctx, c.ID)
	err = errors.Join(err, e)
	for _, a := range allocs {
		_, ad, e := p.allocation(ctx, a.ID)
		dirs, err = append(dirs, ad...), errors.Join(err, e)
	}
	return c, dirs, err
}

func dirsFor(c cluster, user, allocation string) (dirs []dir) {
	r := strings.NewReplacer("{user}", user, "{allocation}", allocation)
	for _, t := range c.Targets {
		uses := func(s string) bool { return strings.Contains(t.Path+t.Owner+t.Group, s) }
		if uses("{user}") == (user != "") && uses("{allocation}") == (allocation != "") {
			dirs = append(dirs, dir{r.Replace(t.Path), r.Replace(t.Owner), r.Replace(t.Group), t.Mode})
		}
	}
	return dirs
}

// ensure runs one SSH session for the plan and fails unless every directory was reported as existing.
func ensure(ctx context.Context, c cluster, dirs []dir) error {
	if len(dirs) == 0 {
		return nil
	}
	args, script, err := command(c, dirs)
	var stderr strings.Builder
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin, cmd.Stderr = strings.NewReader(script), &stderr
	out, runErr := cmd.Output()
	reported := strings.Fields(string(out))
	if missing := slices.DeleteFunc(dirs, func(d dir) bool { return slices.Contains(reported, d.Path) }); len(missing) > 0 {
		err = errors.Join(err, runErr, fmt.Errorf("unix provisioner: %d directories not ensured: %s", len(missing), strings.TrimSpace(stderr.String())))
	}
	return err
}

var valid = regexp.MustCompile(`^(/[A-Za-z0-9_-][A-Za-z0-9_.-]*)+( [a-z0-9_][a-z0-9_.-]*){2} [0-7]{3,4}$`)

// command's script goes on stdin, since a reconcile's outgrows one argument. The leading 00 stops GNU chmod keeping a
// setgid inherited from the parent; every value matches a pattern without shell metacharacters, so none needs quoting.
func command(c cluster, dirs []dir) (args []string, script string, err error) {
	for _, d := range dirs {
		if line := strings.Join([]string{d.Path, d.Owner, d.Group, d.Mode}, " "); !valid.MatchString(line) {
			err = errors.Join(err, fmt.Errorf("unix provisioner: refused %q", line))
		} else {
			script += fmt.Sprintf("test -d %[1]s || { mkdir %[1]s && { chown %[2]s:%[3]s %[1]s && chmod 00%[4]s %[1]s || { rmdir %[1]s; false; }; }; } && echo %[1]s\n", d.Path, d.Owner, d.Group, d.Mode)
		}
	}
	args = []string{"-o", "BatchMode=yes"}
	if c.SSH.Identity != "" {
		args = append(args, "-i", c.SSH.Identity)
	}
	return append(args, c.SSH.User+"@"+c.SSH.Hostname, "sh"), script, err
}
