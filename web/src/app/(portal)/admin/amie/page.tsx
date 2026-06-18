import { redirect } from "next/navigation";

// /admin/amie has no content of its own; bounce to the canonical Inbox tab.
export default function AmieIndex() {
  redirect("/admin/amie/packets");
}
