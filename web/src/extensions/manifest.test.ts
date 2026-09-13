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

import { describe, expect, it } from "vitest";
import {
  extensionRewrites,
  loadConfiguredExtensions,
  validateExtensionManifests,
} from "./manifest";

const manifest = {
  schema_version: 1,
  id: "ssh-certificate-signer",
  name: "SSH Certificate Signer",
  base_path: "/signer",
  web_url: "http://localhost:3001",
  navigation: [
    {
      href: "/signer/certificates",
      label: "SSH Certificates",
      group: "admin",
      icon: "key-round",
      required_privilege: "signer:certificates:read",
    },
  ],
};

describe("extension manifests", () => {
  it("accepts a valid manifest", () => {
    expect(validateExtensionManifests([manifest])).toEqual([manifest]);
  });

  it("generates root and wildcard rewrites for a zone", () => {
    const parsed = validateExtensionManifests([manifest]);
    expect(extensionRewrites(parsed)).toEqual([
      { source: "/signer", destination: "http://localhost:3001/signer" },
      {
        source: "/signer/:path*",
        destination: "http://localhost:3001/signer/:path*",
      },
    ]);
  });

  it("disables discovery when no manifest URLs are configured", async () => {
    await expect(loadConfiguredExtensions("")).resolves.toEqual([]);
  });

  it("fails when a configured manifest is unavailable", async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = async () => new Response(null, { status: 503 });
    await expect(
      loadConfiguredExtensions("https://extensions.example.org/manifest"),
    ).rejects.toThrow(/HTTP 503/);
    globalThis.fetch = originalFetch;
  });

  it("rejects navigation outside the extension base path", () => {
    const invalid = { ...manifest, navigation: [{ ...manifest.navigation[0], href: "/admin" }] };
    expect(() => validateExtensionManifests([invalid])).toThrow(/outside \/signer/);
  });

  it("rejects duplicate ids and base paths", () => {
    expect(() => validateExtensionManifests([manifest, manifest])).toThrow(
      /Duplicate extension id/,
    );
    expect(() =>
      validateExtensionManifests([manifest, { ...manifest, id: "other-extension" }]),
    ).toThrow(/Duplicate extension base path/);
  });

  it("rejects a web URL containing a path", () => {
    expect(() =>
      validateExtensionManifests([{ ...manifest, web_url: "https://example.org/internal" }]),
    ).toThrow(/must be an origin/);
  });

  it("rejects non-http URLs and malformed privilege metadata", async () => {
    expect(() =>
      validateExtensionManifests([{ ...manifest, web_url: "ftp://example.org" }]),
    ).toThrow(/http or https/);
    expect(() =>
      validateExtensionManifests([
        {
          ...manifest,
          navigation: [{ ...manifest.navigation[0], required_privilege: "invalid privilege" }],
        },
      ]),
    ).toThrow(/Invalid extension manifest/);
    await expect(loadConfiguredExtensions("file:///tmp/manifest.json")).rejects.toThrow(
      /http or https/,
    );
  });
});
