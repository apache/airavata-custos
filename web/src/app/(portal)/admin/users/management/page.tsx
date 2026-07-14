import { UsersNav } from "../UsersNav";
import { UsersTableContainer } from "./UsersTableContainer";

export default function UserManagementPage() {
  return (
    <div className="space-y-4">
      <UsersNav />
      <UsersTableContainer />
    </div>
  );
}
