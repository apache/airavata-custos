// Minimal status enums consumed by shared UI primitives.
// TODO(phase-2): promote to full Zod schemas and apiFetch client.

export type AllocationStatus = "ACTIVE" | "INACTIVE" | "DELETED";
export type ChangeRequestStatus = "PENDING" | "APPROVED" | "REJECTED";
