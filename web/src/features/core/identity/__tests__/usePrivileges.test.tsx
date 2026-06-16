import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";

vi.mock("next-auth/react", () => ({
  useSession: () => ({ data: { user: { email: "admin@custos.local" } }, status: "authenticated" }),
}));

import { usePrivileges } from "../queries";

const server = setupServer(
  http.get("*/api/v1/user/privileges", () =>
    HttpResponse.json({ privileges: ["hpc:read", "amie:read"] }),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

function wrapper({ children }: { children: React.ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe("usePrivileges", () => {
  it("fetches and validates the caller privileges payload", async () => {
    const { result } = renderHook(() => usePrivileges(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(["hpc:read", "amie:read"]);
  });
});
