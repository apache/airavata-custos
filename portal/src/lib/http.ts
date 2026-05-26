/**
 * Browser-side JSON fetch helper.
 *
 * All signer API calls go through /api/v1/* (handled by the route at
 * src/app/api/v1/[...path]/route.ts), which is where auth is actually
 * applied. The `auth` option on `RequestOptions` is preserved for parity
 * with the signer API surface but intentionally ignored on the client —
 * forwarding tokens from the browser would leak credentials into the
 * bundle.
 */
type RequestOptions = RequestInit & {
  auth?: "bearer" | "client" | "none";
};

export async function apiFetch<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const { auth = "bearer", headers, ...rest } = options;
  void auth;

  const requestHeaders = new Headers(headers);

  if (!requestHeaders.has("Content-Type") && rest.body) {
    requestHeaders.set("Content-Type", "application/json");
  }

  const response = await fetch(path, {
    ...rest,
    headers: requestHeaders,
  });

  if (!response.ok) {
    let message = `Request failed with ${response.status}`;

    try {
      const body = await response.json();
      message = body.message ?? body.error ?? message;
    } catch {
      // Backend may respond with plain text or empty body on errors.
    }

    throw new Error(message);
  }

  return response.json() as Promise<T>;
}
