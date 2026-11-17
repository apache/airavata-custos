# User Activity connector

Core accepts `POST /me/login-events` only for a verified OIDC caller and queues
`user::portal_login` on the persistent event bus. The portal calls this endpoint
from its initial NextAuth OIDC callback. A stable `jti` + `iat` pair, or `sid` +
`auth_time`, is required to derive a repeatable event ID. The endpoint returns
`202 {"accepted":true}` when queued; the Audit row appears after delivery.

This connector subscribes to that event and inserts one `PORTAL_LOGIN` row in
`audit_events`. The row's deterministic primary key makes retries idempotent.
It exposes the dashboard's `/users/activity`, `/users/activity/analytics`, and
`/users/{id}/activity/analytics` endpoints under
`core:users:activity:read`. All counts are derived from Audit records; there
are no separate login or daily summary tables.

Version one uses **UTC calendar days** for windows, trend buckets, and streaks.
Per-user timezones and retention are separate product decisions. The queries
scan historical Audit login facts; measure them against expected retention and
add a daily projection if dashboard latency requires one.
