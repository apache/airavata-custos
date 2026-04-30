# Zanzibar: Google's Consistent, Global Authorization System — Paper Notes

**Contributor:** Temitope Aderibigbe  
**Source:** USENIX ATC 2019 — Pang et al., Google  
**Georgia Tech VIP Program | Spring 2026**  
**Related Issue:** [AIRAVATA-3978](https://issues.apache.org/jira/browse/AIRAVATA-3978)

---

## 1. What Is Zanzibar?

Zanzibar is Google's globally distributed authorization system — built to answer one core question at massive scale: "Does user U have permission to perform action R on object O?" It is not an authentication system; it is purely an authorization system that determines what an authenticated user is allowed to do.

Zanzibar serves as the single unified access control layer for Google's major products including Calendar, Cloud, Drive, Maps, Photos, and YouTube. Rather than each product building its own permission system, they all delegate authorization decisions to Zanzibar through a shared RPC-based API.

**Scale context:**
- Stores over 2 trillion access control lists (ACLs)
- Handles over 10 million authorization check requests per second
- Maintains 95th-percentile latency under 10 milliseconds
- Sustains 99.999% availability over 3+ years of production use
- Data replicated across 30+ geographic locations worldwide

---

## 2. The Core Data Model — Relation Tuples

Zanzibar's entire authorization model is built on a single primitive: the **relation tuple**. Every permission is expressed as:

object#relation@user

This reads as: "user has relation to object." Examples:

| Tuple | Semantics |
|---|---|
| `doc:readme#owner@10` | User 10 is an owner of doc:readme |
| `group:eng#member@11` | User 11 is a member of group:eng |
| `doc:readme#viewer@group:eng#member` | Members of group:eng are viewers of doc:readme |
| `doc:readme#parent@folder:A` | doc:readme is in folder:A (used for inheritance) |

A key design insight: the `@user` part can itself be an object-relation pair, not just an individual user. This allows ACLs to reference groups and groups to reference other groups — enabling nested group membership and ACL inheritance without storing a separate tuple per object.

Groups are not a special concept in Zanzibar — they are just ACLs where the object is a group and the relation is semantically "member."

---

## 3. Namespace Configuration & Userset Rewrites

Before storing relation tuples, clients define a **namespace configuration** — a logical grouping for a type of object (e.g., `doc`, `folder`, `video`). Each namespace defines its relations and how those relations interact.

### Userset Rewrite Rules

Userset rewrite rules define object-agnostic relationships between relations — for example, "all editors are also viewers." The rule is defined once in the namespace config and applied at evaluation time rather than storing redundant tuples.

Three key primitives:

- **`this`** — Returns all users directly stored for that object-relation pair, including indirect userset references. Default behavior.
- **`computed_userset`** — Derives a userset from another relation on the same object. Example: viewer relation includes the editor userset, so all editors are automatically viewers.
- **`tuple_to_userset`** — Looks up a related object via a tupleset, then computes a userset from that object. This is how folder-level permissions are inherited by documents.

These primitives combine with union, intersection, and exclusion operators to express complex policies compactly.

---

## 4. Consistency Model — The "New Enemy" Problem

Authorization systems face a subtle consistency challenge: if a user is removed from an ACL, they should immediately lose access — even to new content being added around the same time. Zanzibar calls this the **"new enemy" problem**.

**Example A (ACL update order):** Alice removes Bob from a folder ACL, then asks Charlie to add new documents. If the ACL check doesn't respect the ordering, Bob could see the new documents using the stale old ACL.

**Example B (Old ACL on new content):** Alice removes Bob from a document ACL, then Charlie adds new content. If evaluated against a stale snapshot, Bob could see content he was already removed from.

### Zookies — The Consistency Token

Zanzibar solves this with a **zookie** — an opaque consistency token encoding a globally meaningful timestamp:

1. When content is modified, the client requests a zookie via a content-change check. Zanzibar encodes a current global timestamp.
2. The client stores the zookie alongside the content in the same atomic write.
3. Subsequent ACL checks include the zookie — telling Zanzibar to evaluate at a snapshot no older than that timestamp.
4. This guarantees the ACL check always sees state at least as fresh as the content.

The zookie protocol provides at-least-as-fresh semantics without requiring global synchronization on every check. Most checks are served from locally replicated data; only very recent zookies require cross-region round trips.

The underlying storage is Google Spanner, which provides external consistency via its TrueTime mechanism.

---

## 5. The Zanzibar API

| Operation | Description |
|---|---|
| **Read** | Retrieves raw relation tuples matching a tupleset. Does not follow userset rewrite rules. Used to display ACLs or group memberships. |
| **Write** | Adds or removes a single relation tuple. Bulk modifications use read-modify-write with optimistic concurrency control. |
| **Check** | Core operation. Given a userset (`object#relation`), a user, and a zookie — returns whether the user has that relation to the object. |
| **Expand** | Returns the full effective userset for an object-relation pair, following all rewrite rules recursively. Used to build access-controlled search indices. |
| **Watch** | Streams tuple modification events in timestamp order. Used by clients maintaining secondary indices. |

---

## 6. System Architecture

### aclservers
The main server type. Handle Check, Read, Expand, and Write requests. Fan out work to other servers as needed (e.g., recursive group membership traversal). Results gathered and returned to client.

### watchservers
Specialized servers that tail the changelog and stream namespace changes to clients in near real time via the Watch API.

### Spanner
The underlying globally distributed database. Stores relation tuples, namespace configurations, and the changelog. Each namespace gets its own Spanner database. TrueTime provides external consistency guarantees that make zookies meaningful.

### Leopard Indexing System
A specialized index for deeply nested and widely expanded group membership. Standard recursive pointer chasing breaks down for large groups. Leopard addresses this by:

- Maintaining an offline index flattening group-to-group paths (`GROUP2GROUP`) and user-to-group memberships (`MEMBER2GROUP`)
- Answering membership as set intersection: "Is user U a member of group G?" → "Is `MEMBER2GROUP(U) ∩ GROUP2GROUP(G)` non-empty?"
- Maintaining an incremental online layer applying tuple changes in real time on top of the offline snapshot

---

## 7. Performance Techniques

### Distributed Cache
Servers form a distributed cache using consistent hashing. Cache keys encode snapshot timestamps. Timestamp quantization (rounding to 1 or 10 second granularity) allows the vast majority of checks to share the same evaluation snapshot and cache results.

### Hot Spot Mitigation
- Distributed cache with consistent hashing to spread load
- Lock tables to deduplicate simultaneous requests for the same cache key
- Cache prefetching for super-hot objects
- Delayed cancellation of secondary checks when concurrent requests are waiting

### Request Hedging
For Spanner and Leopard calls, Zanzibar sends the same request to multiple servers and uses whichever responds first. Deferred until the initial request is known to be slow to limit extra traffic.

### Performance Isolation
Each client has a global CPU quota. Per-object and per-client concurrency limits on Spanner prevent any single object or client from monopolizing backend resources.

---

## 8. Relevance to Custos & Project 2

Lahiru assigned this paper in the context of the **Dynamic Access Policy & Enforcement Engine (Project 2)** for Custos. The connection is direct:

- **Zanzibar is the canonical example of engine-side authorization** — a dedicated Policy Decision Point (PDP) that applications call at runtime rather than embedding authorization logic in token claims
- **WorkOS FGA**, noted in the platform research as one of the only engine-side systems reviewed, is directly modeled on Zanzibar's relation tuple approach
- **Project 2 calls for** a "policy decision service that other Custos components call to answer: should this user be allowed to do this action on this resource, right now?" — this is precisely the role Zanzibar plays for Google's services
- **Zanzibar's namespace configs** map directly to Project 2's per-tenant policy sets — each Custos tenant (science gateway) defining their own policy set evaluated by a shared engine
- **The audit trail requirement** in Project 2 mirrors Zanzibar's changelog and Watch API, which provide a durable record of all ACL changes for compliance reporting
- **Tools like Open Policy Agent (OPA) and AWS Cedar** are strong candidates as the policy evaluation engine for Project 2, rather than building one from scratch

**Key takeaway:** Zanzibar demonstrates that authorization should be a dedicated service separate from authentication. The question for Custos Project 2 is what subset of Zanzibar's ideas are appropriate for the research computing context and which existing policy engine tools best fit that use case.