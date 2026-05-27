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

  // Read the body once as text so we can both surface non-JSON error
  // messages and avoid the cryptic "Unexpected end of JSON input" that
  // response.json() throws on an empty payload (e.g. 204, or a proxy that
  // strips the body).
  const rawBody = await response.text();
  const parsed = rawBody ? safeParseJson(rawBody) : null;

  if (!response.ok) {
    const message =
      (parsed && typeof parsed === "object"
        ? (parsed as { message?: string; error?: string }).message ??
          (parsed as { message?: string; error?: string }).error
        : undefined) ?? `Request failed with ${response.status}`;
    throw new Error(message);
  }

  return parsed as T;
}

function safeParseJson(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}
