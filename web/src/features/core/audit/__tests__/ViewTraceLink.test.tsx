import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const can = vi.fn<(action: string, subject: string) => boolean>();
const replace = vi.fn();
const push = vi.fn();
let pathnameMock = "/admin/amie/packets";
let searchParamsMock = new URLSearchParams();
let historyReplace: ReturnType<typeof vi.spyOn>;

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can }),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace, push }),
  usePathname: () => pathnameMock,
  useSearchParams: () => searchParamsMock,
}));

beforeEach(() => {
  can.mockReset();
  replace.mockReset();
  push.mockReset();
  can.mockReturnValue(true);
  pathnameMock = "/admin/amie/packets";
  searchParamsMock = new URLSearchParams();
  window.history.replaceState({}, "", "/admin/traces");
  historyReplace = vi.spyOn(window.history, "replaceState");
});

afterEach(() => {
  historyReplace.mockRestore();
  vi.restoreAllMocks();
});

const TRACE_ID = "a3b1c92d3f4e5a6b7c8d9e0f12345678";

describe("ViewTraceLink", () => {
  it("renders nothing when traceId is null", async () => {
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    const { container } = render(<ViewTraceLink traceId={null} />);
    expect(container.firstChild).toBeNull();
  });

  it("renders nothing when ability denies read Trace", async () => {
    can.mockReturnValue(false);
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    const { container } = render(<ViewTraceLink traceId={TRACE_ID} />);
    expect(container.firstChild).toBeNull();
  });

  it("renders an anchor in text variant with 'View trace →'", async () => {
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    render(<ViewTraceLink traceId={TRACE_ID} variant="text" />);
    const link = screen.getByRole("link", { name: /View trace/i });
    expect(link).toBeInTheDocument();
  });

  it("renders icon variant with an aria-label including the short id", async () => {
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    render(<ViewTraceLink traceId={TRACE_ID} variant="icon" />);
    const link = screen.getByRole("link", { name: /View trace a3b1c92d/ });
    expect(link).toBeInTheDocument();
  });

  it("on a non-/admin/traces route, click pushes to /admin/traces/<id>", async () => {
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    render(<ViewTraceLink traceId={TRACE_ID} spanId="1000000000000001" />);

    fireEvent.click(screen.getByRole("link", { name: /View trace/i }));

    expect(push).toHaveBeenCalledWith(
      "/admin/traces/a3b1c92d3f4e5a6b7c8d9e0f12345678?span=1000000000000001",
    );
    expect(replace).not.toHaveBeenCalled();
  });

  it("on /admin/traces, click updates the URL with ?trace=<id> via shallow history", async () => {
    pathnameMock = "/admin/traces";
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    render(<ViewTraceLink traceId={TRACE_ID} />);

    fireEvent.click(screen.getByRole("link", { name: /View trace/i }));

    expect(historyReplace).toHaveBeenCalledTimes(1);
    const target = String(historyReplace.mock.calls[0]?.[2] ?? "");
    expect(target).toBe(`/admin/traces?trace=${TRACE_ID}`);
    expect(push).not.toHaveBeenCalled();
    expect(replace).not.toHaveBeenCalled();
  });

  it("on /admin/traces, click preserves existing filter search params", async () => {
    pathnameMock = "/admin/traces";
    const initialQuery = "status=error&source=amie&from=2026-06-01&q=alloc&limit=50";
    searchParamsMock = new URLSearchParams(initialQuery);
    window.history.replaceState({}, "", `/admin/traces?${initialQuery}`);
    historyReplace.mockClear();
    const { ViewTraceLink } = await import("../components/ViewTraceLink");
    render(<ViewTraceLink traceId={TRACE_ID} spanId="1000000000000001" />);

    fireEvent.click(screen.getByRole("link", { name: /View trace/i }));

    expect(historyReplace).toHaveBeenCalledTimes(1);
    const target = String(historyReplace.mock.calls[0]?.[2] ?? "");
    const url = new URL(target, "http://localhost");
    expect(url.pathname).toBe("/admin/traces");
    expect(url.searchParams.get("status")).toBe("error");
    expect(url.searchParams.get("source")).toBe("amie");
    expect(url.searchParams.get("from")).toBe("2026-06-01");
    expect(url.searchParams.get("q")).toBe("alloc");
    expect(url.searchParams.get("limit")).toBe("50");
    expect(url.searchParams.get("trace")).toBe(TRACE_ID);
    expect(url.searchParams.get("span")).toBe("1000000000000001");
  });
});
