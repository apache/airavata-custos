import { Button } from "@/shared/ui/button";
import { UsersNav } from "../UsersNav";
import { RoleFormDialog } from "./RoleFormDialog";
import { RolesGrid } from "./RolesGrid";

export default function RoleManagementPage() {
  return (
    <div className="space-y-6">
      <UsersNav
        rightSlot={<RoleFormDialog triggerRender={<Button />} triggerContent="Create role" />}
      />
      <RolesGrid />
    </div>
  );
}
