type RequestOptions = RequestInit & {
  auth?: "bearer" | "client" | "none";
};

export async function apiFetch<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  // Auth is applied by the Next API proxy; keeping the option preserves the
  // signer API call shape without exposing credentials in browser code.
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
      // Some backend errors may be plain text or empty responses.
    }

    throw new Error(message);
  }

  return response.json() as Promise<T>;
}
