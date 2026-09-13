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

import { z } from "zod";

const navigationSchema = z.object({
  href: z.string().regex(/^\/[a-z0-9]+(?:-[a-z0-9]+)*(?:\/[a-z0-9]+(?:-[a-z0-9]+)*)*$/),
  label: z.string().min(1),
  group: z.enum(["allocations", "admin"]),
  icon: z.enum(["key-round"]),
  required_privilege: z.string().regex(/^[a-z0-9-]+(?::[a-z0-9-]+)+$/),
});

const manifestSchema = z.object({
  schema_version: z.literal(1),
  id: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/),
  name: z.string().min(1),
  base_path: z.string().regex(/^\/[a-z0-9]+(?:-[a-z0-9]+)*(?:\/[a-z0-9]+(?:-[a-z0-9]+)*)*$/),
  web_url: z.string().url(),
  navigation: z.array(navigationSchema).min(1),
});

export type ExtensionManifest = z.infer<typeof manifestSchema>;

function httpURL(value: string, label: string): URL {
  let url: URL;
  try {
    url = new URL(value);
  } catch {
    throw new Error(`${label} must be a valid URL`);
  }
  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error(`${label} must use http or https`);
  }
  return url;
}

export function validateExtensionManifests(values: unknown[]): ExtensionManifest[] {
  const manifests = values.map((value, index) => {
    const parsed = manifestSchema.safeParse(value);
    if (!parsed.success) {
      throw new Error(`Invalid extension manifest at index ${index}: ${parsed.error.message}`);
    }
    const manifest = parsed.data;
    const webURL = httpURL(manifest.web_url, `Extension ${manifest.id} web_url`);
    if (webURL.pathname !== "/" || webURL.search || webURL.hash) {
      throw new Error(`Extension ${manifest.id} web_url must be an origin without a path`);
    }
    if (
      manifest.navigation.some(
        (item) =>
          item.href !== manifest.base_path && !item.href.startsWith(`${manifest.base_path}/`),
      )
    ) {
      throw new Error(`Extension ${manifest.id} has navigation outside ${manifest.base_path}`);
    }
    return manifest;
  });

  const ids = new Set<string>();
  const basePaths = new Set<string>();
  for (const manifest of manifests) {
    if (ids.has(manifest.id)) throw new Error(`Duplicate extension id: ${manifest.id}`);
    if (basePaths.has(manifest.base_path)) {
      throw new Error(`Duplicate extension base path: ${manifest.base_path}`);
    }
    ids.add(manifest.id);
    basePaths.add(manifest.base_path);
  }
  return manifests;
}

export async function loadConfiguredExtensions(
  configured = process.env.CUSTOS_EXTENSION_MANIFEST_URLS ?? "",
): Promise<ExtensionManifest[]> {
  const urls = configured
    .split(",")
    .map((value) => value.trim())
    .filter(Boolean);
  for (const url of urls) httpURL(url, "Extension manifest URL");
  const values = await Promise.all(
    urls.map(async (url) => {
      const response = await fetch(url, { cache: "no-store" });
      if (!response.ok) {
        throw new Error(`Extension manifest ${url} returned HTTP ${response.status}`);
      }
      return response.json() as Promise<unknown>;
    }),
  );
  return validateExtensionManifests(values);
}

export function extensionRewrites(manifests: ExtensionManifest[]) {
  return manifests.flatMap((manifest) => {
    const webURL = manifest.web_url.replace(/\/$/, "");
    return [
      { source: manifest.base_path, destination: `${webURL}${manifest.base_path}` },
      {
        source: `${manifest.base_path}/:path*`,
        destination: `${webURL}${manifest.base_path}/:path*`,
      },
    ];
  });
}
