import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { groupPrivileges, PrivilegeList } from "../PrivilegeList";

describe("PrivilegeList", () => {
  it("groups real read/write keys without inventing write", () => {
    expect(groupPrivileges(["core:traces:read", "core:users:read", "core:users:write"])).toEqual([
      { kind: "rw", key: "core:traces", read: true, write: false },
      { kind: "rw", key: "core:users", read: true, write: true },
    ]);
  });

  it("preserves meta and unknown privileges verbatim", () => {
    render(<PrivilegeList privileges={["core:roles:manage", "connector:custom:inspect"]} />);
    expect(screen.getByText("core:roles:manage")).toBeInTheDocument();
    expect(screen.getByText("connector:custom:inspect")).toBeInTheDocument();
  });

  it("deduplicates repeated keys", () => {
    render(<PrivilegeList privileges={["core:roles:manage", "core:roles:manage"]} />);
    expect(screen.getAllByText("core:roles:manage")).toHaveLength(1);
  });
});
