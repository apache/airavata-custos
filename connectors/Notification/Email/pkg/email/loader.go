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

// Package email is the email notification connector entry point. Wired from
// internal/connectors/loader.go.
package email

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/Notification/Email/internal/notify"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "email-notifier"

// LoadConnector registers the email subscribers.
func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *identity.Router, cfg *config.ConnectorConfig) error {
	get := func(key string) string {
		v, _ := cfg.GetStringField(key)
		// A "${...}" left in place is an env var that was not set.
		if strings.HasPrefix(v, "${") {
			return ""
		}
		return v
	}
	port, err := cfg.GetIntField("smtp_port")
	if err != nil {
		port = 587
	}
	host, username, password, from := get("smtp_host"), get("username"), get("password"), get("from")
	site := notify.Site{
		SiteName:     get("site_name"),
		PortalURL:    get("portal_url"),
		ClusterHost:  get("cluster_host"),
		SupportEmail: get("support_email"),
		LogoURL:      get("logo_url"),
	}
	for key, v := range map[string]string{
		"smtp_host": host, "username": username, "password": password, "from": from,
		"site_name": site.SiteName, "portal_url": site.PortalURL, "cluster_host": site.ClusterHost, "support_email": site.SupportEmail,
	} {
		if v == "" {
			slog.Info("email notifications: required setting not set; skipping", "setting", key)
			return nil
		}
	}

	if site.LogoURL == "" {
		site.LogoURL = strings.TrimSuffix(site.PortalURL, "/") + "/brand/logo/custos-mark.png"
	}

	sender, err := notify.NewSMTP(host, port, username, password, from)
	if err != nil {
		return err
	}
	notify.NewNotifier(coreService, sender, site).RegisterSubscribers(eventBus, Type)
	slog.Info("email notifications: subscribers registered", "smtp_host", host, "from", from)
	return nil
}
