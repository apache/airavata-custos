export function responseBodyForStatus(status: number, body: string): string | null {
  // The Fetch standard forbids a body for these statuses. Passing even an
  // empty string makes the Response constructor throw, turning a successful
  // upstream DELETE (204) into a portal-side 500.
  return status === 204 || status === 205 || status === 304 ? null : body;
}
