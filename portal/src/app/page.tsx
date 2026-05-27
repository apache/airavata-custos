// Portal root route ("/"). The dashboard is not implemented yet and the
// only feature surface is /signer/certificates, so redirect there. Using
// next/navigation's redirect() turns this into a clean server-side 307
// instead of rendering anything, which also sidesteps any client-side
// session/json parsing on the root render path.
import { redirect } from "next/navigation";

export default function HomePage() {
  redirect("/signer/certificates");
}
