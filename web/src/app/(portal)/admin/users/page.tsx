import { redirect } from "next/navigation";

// /admin/users has no content of its own; bounce to the canonical tab.
export default function AdminUsersIndex() {
  redirect("/admin/users/management");
}
