// Custom Auth.js error route registered via `pages.error` in auth.ts.
// Lives at /auth-error (not /api/auth/error) so a refresh re-renders a
// normal portal page where the user can retry, rather than re-rendering
// Auth.js's stock error screen. Interactive bits live in AuthErrorView so
// this route can stay a Server Component and provide the Suspense boundary
// required for `useSearchParams` under the App Router.
import { Suspense } from "react";
import { AuthErrorView } from "./AuthErrorView";

export default function AuthErrorPage() {
  return (
    <Suspense>
      <AuthErrorView />
    </Suspense>
  );
}
