// Custom Auth.js sign-in route registered via `pages.signIn` in auth.ts.
// Renders branded portal chrome instead of Auth.js's stock provider list.
// The interactive bits live in SignInForm so this route can stay a Server
// Component and use the Suspense boundary required when reading search
// params under the App Router.
import { Suspense } from "react";
import { SignInForm } from "./SignInForm";

export default function SignInPage() {
  return (
    <Suspense>
      <SignInForm />
    </Suspense>
  );
}
