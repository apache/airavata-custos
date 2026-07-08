import { UsersNav } from "../UsersNav";
import { UsersTable } from "./UsersTable";

export default function UserManagementPage() {
  return (
    <div className="space-y-4">
      <UsersNav />
      <UsersTable />
    </div>
  );
}
