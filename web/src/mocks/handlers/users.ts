// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0.

import { http, HttpResponse } from "msw";
import fixture from "@/features/core/identity/__fixtures__/settings.json";

const roles = Object.values(fixture.roleDetails).map((detail) => detail.role);

export const usersHandlers = [
  http.get("*/api/v1/users", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? 25);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const users = [fixture.user];
    return HttpResponse.json({
      items: users.slice(offset, offset + limit),
      total: users.length,
    });
  }),
  http.get("*/api/v1/roles", () => HttpResponse.json(roles)),
  http.post("*/api/v1/users/:id/roles", async ({ params, request }) => {
    const body = (await request.json()) as { role_id?: string };
    return HttpResponse.json(
      {
        user_id: String(params.id),
        role_id: body.role_id,
        granted_at: new Date().toISOString(),
      },
      { status: 201 },
    );
  }),
  http.delete("*/api/v1/users/:id/roles/:roleId", () => new HttpResponse(null, { status: 204 })),
];
