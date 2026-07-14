import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { RoleAssignMenu } from "../RoleAssignMenu";

const roleDetailsMock = vi.hoisted(() => ({
  roles: [
    {
      id: "role-1",
      name: "Administrator",
      privileges: ["core:roles:manage"],
    },
  ],
  isLoading: false,
  isError: false,
}));

vi.mock("@/features/core/users/queries", () => ({
  useRoleDetails: () => roleDetailsMock,
}));

describe("RoleAssignMenu", () => {
  beforeEach(() => {
    roleDetailsMock.roles = [
      {
        id: "role-1",
        name: "Administrator",
        privileges: ["core:roles:manage"],
      },
    ];
    roleDetailsMock.isLoading = false;
    roleDetailsMock.isError = false;
  });

  it("submits the desired role set once", async () => {
    const onSave = vi.fn().mockResolvedValue(true);
    render(
      <RoleAssignMenu
        roles={[{ id: "role-1", name: "Administrator" }]}
        heldRoleIds={new Set(["role-1"])}
        onSave={onSave}
        triggerLabel="Manage user roles"
        isCurrentUser={false}
        isPending={false}
        error={null}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /edit roles/i }));
    fireEvent.click(screen.getByRole("button", { name: "Unassign" }));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith([]));
    expect(onSave).toHaveBeenCalledTimes(1);
  });

  it("confirms before self-removing a role-manager role", async () => {
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(false);
    const onSave = vi.fn().mockResolvedValue(true);
    render(
      <RoleAssignMenu
        roles={[{ id: "role-1", name: "Administrator" }]}
        heldRoleIds={new Set(["role-1"])}
        onSave={onSave}
        triggerLabel="Manage user roles"
        isCurrentUser
        isPending={false}
        error={null}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /edit roles/i }));
    fireEvent.click(screen.getByRole("button", { name: "Unassign" }));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(confirm).toHaveBeenCalled();
    expect(onSave).not.toHaveBeenCalled();
    confirm.mockRestore();
  });

  it("allows saving when role privilege previews are unavailable", async () => {
    roleDetailsMock.roles = [];
    roleDetailsMock.isError = true;
    const onSave = vi.fn().mockResolvedValue(true);
    render(
      <RoleAssignMenu
        roles={[{ id: "role-1", name: "Administrator" }]}
        heldRoleIds={new Set(["role-1"])}
        onSave={onSave}
        triggerLabel="Manage user roles"
        isCurrentUser={false}
        isPending={false}
        error={null}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /edit roles/i }));
    fireEvent.click(screen.getByRole("button", { name: "Unassign" }));
    const save = screen.getByRole("button", { name: "Save" });
    expect(save).toBeEnabled();
    fireEvent.click(save);

    await waitFor(() => expect(onSave).toHaveBeenCalledWith([]));
  });

  it("uses a conservative confirmation when current-user privileges are unavailable", () => {
    roleDetailsMock.roles = [];
    roleDetailsMock.isError = true;
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(false);
    const onSave = vi.fn().mockResolvedValue(true);
    render(
      <RoleAssignMenu
        roles={[{ id: "role-1", name: "Administrator" }]}
        heldRoleIds={new Set(["role-1"])}
        onSave={onSave}
        triggerLabel="Manage user roles"
        isCurrentUser
        isPending={false}
        error={null}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /edit roles/i }));
    fireEvent.click(screen.getByRole("button", { name: "Unassign" }));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(confirm).toHaveBeenCalledWith(expect.stringContaining("privileges are unavailable"));
    expect(onSave).not.toHaveBeenCalled();
    confirm.mockRestore();
  });
});
