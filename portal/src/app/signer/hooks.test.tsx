/**
 * Vitest coverage for the signer data hooks. Mocks the global fetch so the
 * hooks can be exercised without a live signer or NextAuth session, and
 * verifies that each hook fires the expected /api/v1 call and surfaces
 * success/error state through its returned object.
 *
 * @vitest-environment jsdom
 */
import React from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act } from "react";
import { createRoot } from "react-dom/client";
import {
  useCertificate,
  useCertificates,
  useRevokeCertificate,
  useUserInfo,
} from "./hooks";
import type { Certificate } from "./types";

(globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT: boolean })
  .IS_REACT_ACT_ENVIRONMENT = true;

type HookResult<T> = {
  current: () => T;
  unmount: () => void;
};

const sampleCertificate: Certificate = {
  serial_number: 42,
  key_id: "researcher@login.cluster.example.org",
  principal: "researcher",
  public_key_fingerprint: "SHA256:sample",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 1_800_000_000,
  issued_at: 1_700_000_000,
  revoked: false,
};

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("signer data hooks", () => {
  it("loads current user info", async () => {
    fetchMock().mockResolvedValueOnce(jsonResponse({
      subject: "sub",
      issuer: "issuer",
      email: "researcher@example.org",
      principal: "researcher",
    }));

    const hook = await renderHook(() => useUserInfo());

    await waitFor(() => expect(hook.current().loading).toBe(false));
    expect(hook.current().data?.principal).toBe("researcher");
    expect(fetchMock()).toHaveBeenCalledWith("/api/v1/userinfo", {
      headers: expect.any(Headers),
    });

    hook.unmount();
  });

  it("reports user info loading errors", async () => {
    fetchMock().mockResolvedValueOnce(
      jsonResponse({ message: "not authenticated" }, false, 401)
    );

    const hook = await renderHook(() => useUserInfo());

    await waitFor(() => expect(hook.current().loading).toBe(false));
    expect(hook.current().error).toBe("not authenticated");

    hook.unmount();
  });

  it("loads certificates with paging and admin username filters", async () => {
    fetchMock().mockResolvedValueOnce(
      jsonResponse({
        certificates: [sampleCertificate],
        total: 1,
        limit: 10,
        offset: 5,
      })
    );

    const hook = await renderHook(() => useCertificates(10, 5, "researcher"));

    await waitFor(() => expect(hook.current().loading).toBe(false));
    expect(hook.current().data?.certificates).toHaveLength(1);
    expect(fetchMock().mock.calls[0][0]).toBe(
      "/api/v1/certificates?limit=10&offset=5&username=researcher"
    );

    hook.unmount();
  });

  it("loads certificate details by serial", async () => {
    fetchMock().mockResolvedValueOnce(jsonResponse(sampleCertificate));

    const hook = await renderHook(() => useCertificate("42"));

    await waitFor(() => expect(hook.current().loading).toBe(false));
    expect(hook.current().data?.serial_number).toBe(42);
    expect(fetchMock().mock.calls[0][0]).toBe("/api/v1/certificates/42");

    hook.unmount();
  });

  it("reports certificate detail loading errors", async () => {
    fetchMock().mockResolvedValueOnce(
      jsonResponse({ error: "certificate not found" }, false, 404)
    );

    const hook = await renderHook(() => useCertificate("404"));

    await waitFor(() => expect(hook.current().loading).toBe(false));
    expect(hook.current().error).toBe("certificate not found");

    hook.unmount();
  });

  it("posts revoke requests and exposes mutation state", async () => {
    fetchMock().mockResolvedValueOnce(
      jsonResponse({
        success: true,
        message: "revoked",
        revoked_count: 1,
      })
    );

    const hook = await renderHook(() => useRevokeCertificate());

    await act(async () => {
      await hook.current().revoke({
        serial_number: 42,
        reason: "No longer needed",
      });
    });

    expect(hook.current().loading).toBe(false);
    expect(fetchMock().mock.calls[0][0]).toBe("/api/v1/revoke");
    expect(fetchMock().mock.calls[0][1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({
        serial_number: 42,
        reason: "No longer needed",
      }),
    });

    hook.unmount();
  });
});

async function renderHook<T>(useHook: () => T): Promise<HookResult<T>> {
  const container = document.createElement("div");
  const root = createRoot(container);
  let value: T;

  function TestComponent() {
    value = useHook();
    return null;
  }

  await act(async () => {
    root.render(<TestComponent />);
  });

  return {
    current: () => value,
    unmount: () => {
      act(() => {
        root.unmount();
      });
    },
  };
}

function fetchMock() {
  return fetch as unknown as ReturnType<typeof vi.fn>;
}

function jsonResponse(body: unknown, ok = true, status = 200): Response {
  // Mirror a real Response: apiFetch now reads via text() so an empty body
  // can be detected before JSON.parse.
  return {
    ok,
    status,
    json: async () => body,
    text: async () => JSON.stringify(body),
  } as Response;
}

async function waitFor(assertion: () => void) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < 1000) {
    try {
      assertion();
      return;
    } catch {
      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 10));
      });
    }
  }

  assertion();
}
