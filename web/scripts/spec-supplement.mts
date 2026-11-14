/**
 * What the portal states about the API beyond master's swag specs, read off master's handlers.
 * Responses a handler writes as a map literal carry every key, so each is listed with the Go value
 * behind it: nullable only where a nil slice or pointer can reach the encoder. A changed key set
 * fails generation.
 */
import { type GoIndex, type Node, at, isNode } from "./go-presence.mts";

const ref = (name: string) => ({ $ref: `#/definitions/${name}` });
const array = (items: unknown) => ({ type: "array", items });
const nullable = (schema: Node) => ({ ...schema, "x-nullable": true });
const str = { type: "string" };
const int = { type: "integer" };
const bool = { type: "boolean" };
const literal = (properties: Node) => ({
  type: "object",
  properties,
  required: Object.keys(properties),
});

// Keyed by gen-api module, then by operation.
const inline: Record<string, Record<string, (go: GoIndex) => Node>> = {
  core: {
    // A nil event list is replaced with an empty one.
    "GET /audit/events": () => literal({ events: array(ref("TraceEvent")) }),
    // The store answers a fixed list.
    "GET /audit/sources": () => literal({ sources: array(str) }),
    // ListTraces builds its rows with make.
    "GET /audit/traces": () =>
      literal({ traces: array(ref("TraceSummary")), total: int, limit: int, offset: int }),
    // A trace with rows hangs each under the root or a parent, and nil deliveries become empty.
    "GET /audit/traces/{trace_id}": () =>
      literal({
        trace_id: str,
        status: str,
        tree: array(ref("TraceNode")),
        deliveries: array(ref("TraceDelivery")),
        truncated: bool,
      }),
    // ListDeliveries selects into a non-nil slice.
    "GET /events/deliveries": () => literal({ items: array(ref("PendingDelivery")) }),
    // ListSubscriptions builds its rows with make.
    "GET /events/subscriptions": () => literal({ items: array(ref("Subscription")) }),
    // GetRole answers not found rather than a nil role; the privilege select may stay nil.
    "GET /roles/{id}": () =>
      literal({ role: ref("Role"), privileges: nullable(array(ref("PrivilegeKey"))) }),
    // A nil privilege set is replaced with an empty one.
    "GET /me/privileges": () => literal({ privileges: array(ref("PrivilegeKey")) }),
  },
  amie: {
    // ListAuditsForPacket builds models.TraceEvent rows with make, id included.
    "GET /connectors/amie/packets/{packet_id}/audits": (go) =>
      literal({ packet_id: str, events: array(struct(go, "TraceEvent", traceEventProps)) }),
  },
};

// List routes whose annotation names another model than the handler returns: [annotated, returned].
const retyped: Record<string, Record<string, [string, string]>> = {
  core: {
    // ListResourcesForAllocation returns the attached resources, not their mappings.
    "GET /compute-allocations/{id}/resources": [
      "ComputeAllocationResourceMapping",
      "ComputeAllocationResource",
    ],
  },
};

// Slice and map fields a handler never leaves nil, though their Go type could be: [definition, field].
const filled: Record<string, [string, string][]> = {
  core: [
    // user.go and organization.go replace a nil page with an empty one.
    ["UserListResponse", "items"],
    ["OrganizationListResponse", "items"],
    // allocation.go replaces a nil page with an empty one; project.go and compute_cluster.go build theirs with make.
    ["ComputeAllocationListResponse", "items"],
    ["ProjectListResponse", "items"],
    ["ClusterAccountListResponse", "items"],
    // FindByProjectWithUser coalesces the aggregate to '[]'.
    ["ProjectMemberResponse", "allocations"],
    // EffectivePrivileges and getCallerProfile build theirs with make; ListDetailedByUser starts each at [].
    ["CallerProfileResponse", "privileges"],
    ["CallerProfileResponse", "roles"],
    ["CallerRoleGrant", "privileges"],
  ],
  amie: [
    // Each handler builds its page with make or a literal.
    ["PacketListResponse", "packets"],
    ["PacketStatsResponse", "byDay"],
    ["ReplyListResponse", "replies"],
  ],
  analytics: [
    // The service builds every list with make and fills each day's map.
    ["AllocationJobs", "jobs"],
    ["ProjectContext", "allocations"],
    ["UsageSummary", "daily"],
    ["UsageSummary", "by_resource"],
    ["UsageDailyBucket", "by_resource"],
  ],
};

// List routes answering a store select's slice as is, so an empty result encodes as null.
const nilLists: Record<string, string[]> = {
  core: [
    "GET /roles",
    "GET /roles/{id}/holders",
    "GET /users/{id}/roles",
    "GET /users/{id}/privileges",
    "GET /users/{id}/user-identities",
    "GET /compute-clusters",
    "GET /compute-clusters/{id}/users",
    "GET /compute-allocation-resources/summary",
    "GET /compute-allocation-resources/{id}/rates",
  ],
};

const traceEventProps = Object.fromEntries(
  [
    "id",
    "span_id",
    "parent_span_id",
    "source",
    "event_type",
    "entity_type",
    "entity_id",
    "description",
    "status",
    "created_at",
  ].map((k) => [k, str]),
);

function struct(go: GoIndex, name: string, properties: Node): Node {
  const rules = go.rules(name, Object.keys(properties));
  if (!rules) throw new Error(`spec-supplement: no Go struct ${name} matches its properties`);
  return {
    type: "object",
    properties,
    required: [...rules].filter(([, p]) => p.required).map(([k]) => k),
  };
}

const success = (spec: Node, op: string) => {
  const [method = "", route = ""] = op.split(" ");
  return at(spec, "paths", route, method.toLowerCase(), "responses", "200");
};

/** Settles nullability the Go types cannot tell, after the presence pass, in place. */
export function settle(module: string, spec: Node): void {
  for (const [name, field] of filled[module] ?? []) {
    const properties = at(spec, "definitions", name, "properties");
    const prop = at(properties, field);
    const { "x-nullable": wasNullable, ...rest } = isNode(prop) ? prop : {};
    if (
      !isNode(properties) ||
      wasNullable !== true ||
      !["array", "object"].includes(String(at(rest, "type")))
    ) {
      throw new Error(`spec-supplement: ${name}.${field} is no longer a nullable list or map`);
    }
    properties[field] = rest;
  }
  for (const op of nilLists[module] ?? []) {
    const response = success(spec, op);
    const schema = at(response, "schema");
    if (!isNode(response) || !isNode(schema) || at(schema, "type") !== "array") {
      throw new Error(`spec-supplement: ${op} no longer answers a list`);
    }
    Object.assign(response, { schema: nullable(schema) });
  }
}

/** Applies the supplement for one gen-api module to its parsed spec, in place. */
export function supplement(module: string, spec: Node, go: GoIndex): void {
  for (const [op, [annotated, returned]] of Object.entries(retyped[module] ?? {})) {
    const response = success(spec, op);
    if (!isNode(response) || at(response, "schema", "items", "$ref") !== ref(annotated).$ref) {
      throw new Error(`spec-supplement: ${op} no longer annotates ${annotated}[]`);
    }
    Object.assign(response, { schema: array(ref(returned)) });
  }
  for (const [op, schema] of Object.entries(inline[module] ?? {})) {
    const response = success(spec, op);
    const stated = schema(go);
    const keys = (s: unknown) => {
      const properties = at(s, "properties");
      return Object.keys(isNode(properties) ? properties : {})
        .sort()
        .join();
    };
    if (!isNode(response) || keys(at(response, "schema")) !== keys(stated)) {
      throw new Error(`spec-supplement: ${op} no longer answers ${keys(stated)}`);
    }
    Object.assign(response, { schema: stated });
  }
}
