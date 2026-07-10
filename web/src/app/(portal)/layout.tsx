import { redirect } from "next/navigation";
import { auth } from "@/shared/auth/auth";
import { PortalLayout } from "@/shared/layout/PortalLayout";
import { UsersAdminProvider } from "@/shared/users-admin/UsersAdminContext";

export default async function PortalRoutesLayout({ children }: { children: React.ReactNode }) {
  const session = await auth();
  if (!session?.user) redirect("/sign-in");
  if ((session.privileges ?? []).length === 0) redirect("/no-access");
  return (
    <UsersAdminProvider>
      <PortalLayout>{children}</PortalLayout>
    </UsersAdminProvider>
  );
}
