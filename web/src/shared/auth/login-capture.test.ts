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

import { describe, expect, it, vi } from "vitest";
import { captureInitialOIDCLogin, captureLoginEvent, loginCaptureBearer } from "./login-capture";

describe("captureInitialOIDCLogin", () => {
  it("initiates capture exactly once for an initial account callback", async () => {
    const capture = vi.fn().mockResolvedValue(true);

    const result = await captureInitialOIDCLogin({
      account: { access_token: "header.payload.signature", id_token: "id-token" },
      coreApiBaseUrl: "http://core",
      capture,
    });

    expect(result).toBe(true);
    expect(capture).toHaveBeenCalledTimes(1);
    expect(capture).toHaveBeenCalledWith({
      bearer: "header.payload.signature",
      coreApiBaseUrl: "http://core",
      fetcher: fetch,
    });
  });

  it("does not capture for session reads without account data", async () => {
    const capture = vi.fn();

    const result = await captureInitialOIDCLogin({
      coreApiBaseUrl: "http://core",
      capture,
    });

    expect(result).toBeUndefined();
    expect(capture).not.toHaveBeenCalled();
  });

  it("uses a JWT-shaped id token when the access token is opaque", () => {
    expect(
      loginCaptureBearer({ access_token: "opaque", id_token: "header.payload.signature" }),
    ).toBe("header.payload.signature");
  });
});

describe("captureLoginEvent", () => {
  it("submits once when verified OIDC session claims are present", async () => {
    const fetcher = vi.fn().mockResolvedValue(new Response(null, { status: 201 }));
    const sleeper = vi.fn();

    await captureLoginEvent({
      bearer: "verified-bearer",
      coreApiBaseUrl: "http://core",
      fetcher,
      sleeper,
    });

    expect(fetcher).toHaveBeenCalledTimes(1);
    expect(sleeper).not.toHaveBeenCalled();
    expect(fetcher).toHaveBeenCalledWith(
      "http://core/me/login-events",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("retries transient failures with the same bearer and empty evidence payload", async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 503 }))
      .mockRejectedValueOnce(new Error("network"))
      .mockResolvedValueOnce(new Response(null, { status: 200 }));
    const sleeper = vi.fn().mockResolvedValue(undefined);

    await captureLoginEvent({
      bearer: "verified-bearer",
      coreApiBaseUrl: "http://core",
      fetcher,
      sleeper,
    });

    expect(fetcher).toHaveBeenCalledTimes(3);
    const bodies = fetcher.mock.calls.map((call) => call[1]?.body);
    expect(new Set(bodies).size).toBe(1);
    expect(sleeper).toHaveBeenCalledTimes(2);
  });

  it("does not retry permanent input failures", async () => {
    const fetcher = vi.fn().mockResolvedValue(new Response(null, { status: 400 }));
    await captureLoginEvent({
      bearer: "verified-bearer",
      coreApiBaseUrl: "http://core",
      fetcher,
      sleeper: vi.fn(),
    });
    expect(fetcher).toHaveBeenCalledTimes(1);
  });
  it("bounds each request and allows login to continue after exhausted failures", async () => {
    const fetcher = vi.fn().mockRejectedValue(new Error("unavailable"));
    const sleeper = vi.fn().mockResolvedValue(undefined);
    const result = await captureLoginEvent({
      bearer: "test-token",
      coreApiBaseUrl: "http://core",
      fetcher,
      sleeper,
    });
    expect(result).toBe(false);
    expect(fetcher).toHaveBeenCalledTimes(3);
    expect(fetcher.mock.calls[0]?.[1]).toEqual(
      expect.objectContaining({ body: "{}", signal: expect.any(AbortSignal) }),
    );
    expect(sleeper.mock.calls.map((call) => call[0])).toEqual([50, 100]);
  });
});
