import { http, HttpResponse } from "msw";

export const healthzHandlers = [
  http.get("/api/v1/healthz", () => HttpResponse.json({ status: "ok" })),
];
