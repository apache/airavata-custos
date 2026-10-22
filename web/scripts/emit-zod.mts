/**
 * Writes zod.gen.ts and types.gen.ts for one settled Swagger 2.0 document under the names the portal
 * imports: zX and X per definition, and zOpBody, zOpPath, zOpQuery, zOpResponse and OpData per
 * operation, Op being the method and path in PascalCase (`GET /users/{id}` → GetUsersById). Names
 * keep hey-api's casing, which splits acronyms (SU → Su). The renderer covers only the schema shapes
 * swag emits and throws on anything else; a $ref becomes the referenced constant, lazily only when it
 * closes a cycle.
 */
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { type Node, isNode } from "./go-presence.mjs";

type Schema = {
  $ref?: string;
  enum?: unknown[];
  type?: string;
  items?: Schema;
  properties?: Record<string, Schema>;
  additionalProperties?: Schema | boolean;
  required?: string[];
  allOf?: Schema[];
  "x-nullable"?: boolean;
};
type Param = Schema & { in: string; name: string; required?: boolean; schema?: Schema };
type Operation = { parameters?: Param[]; responses?: Record<string, { schema?: Schema }> };
type Spec = { definitions?: Record<string, Schema>; paths?: Record<string, Record<string, Operation>> };

const words = (s: string) => s.match(/[A-Z]+(?![a-z])|[A-Z]?[a-z]+|[0-9]+/g) ?? [];
const pascal = (s: string) =>
  words(s)
    .map((w) => w[0]?.toUpperCase() + w.slice(1).toLowerCase())
    .join("");
const defName = (ref: string) => ref.replace("#/definitions/", "");

function refs(value: unknown, out = new Set<string>()): Set<string> {
  if (Array.isArray(value)) for (const v of value) refs(v, out);
  else if (isNode(value)) {
    const { $ref } = value as Schema;
    if ($ref) out.add(defName($ref));
    for (const v of Object.values(value)) refs(v, out);
  }
  return out;
}

function base(s: Schema, emitted: Set<string>): string {
  if (s.allOf?.length === 1 && s.allOf[0]) return base(s.allOf[0], emitted);
  if (s.$ref) {
    const id = `z${pascal(defName(s.$ref))}`;
    return emitted.has(defName(s.$ref)) ? id : `z.lazy((): any => ${id})`;
  }
  if (s.enum) return `z.enum(${JSON.stringify(s.enum)})`;
  switch (s.type) {
    case undefined:
      if (!s.properties && !s.allOf) return "z.unknown()";
      break;
    case "string":
      return "z.string()";
    case "integer":
      return "z.int()";
    case "number":
      return "z.number()";
    case "boolean":
      return "z.boolean()";
    case "array":
      return `z.array(${render(s.items, emitted)})`;
    case "object": {
      if (isNode(s.additionalProperties)) {
        return `z.record(z.string(), ${render(s.additionalProperties, emitted)})`;
      }
      if (!s.properties) return "z.record(z.string(), z.unknown())";
      const required = new Set(s.required);
      const fields = Object.entries(s.properties).map(
        ([k, v]) => `${JSON.stringify(k)}: ${render(v, emitted)}${required.has(k) ? "" : ".optional()"}`,
      );
      return `z.object({ ${fields.join(", ")} })`;
    }
  }
  throw new Error(`emit-zod: unsupported schema ${JSON.stringify(s)}`);
}

function render(s: Schema | undefined, emitted: Set<string>): string {
  if (!s) throw new Error("emit-zod: missing schema");
  return base(s, emitted) + (s["x-nullable"] ? ".nullable()" : "");
}

const object = (params: Param[]): Schema => ({
  type: "object",
  properties: Object.fromEntries(
    params.map(({ in: _, name, required: __, ...schema }) => [name, schema]),
  ),
  required: params.filter((p) => p.required).map((p) => p.name),
});

export function emit(document: Node, output: string): void {
  const { definitions = {}, paths = {} } = document as Spec;
  const emitted = new Set<string>();
  const visiting = new Set<string>();
  const zod = ['import { z } from "zod";', ""];
  const types = ['import type { z } from "zod";', 'import type * as s from "./zod.gen";', ""];

  const define = (name: string) => {
    if (emitted.has(name) || visiting.has(name) || !definitions[name]) return;
    visiting.add(name);
    for (const dep of refs(definitions[name])) define(dep);
    const id = pascal(name);
    zod.push(`export const z${id} = ${render(definitions[name], emitted)};`);
    zod.push(`export type ${id} = z.infer<typeof z${id}>;`, "");
    emitted.add(name);
  };
  for (const name of Object.keys(definitions).sort()) define(name);

  for (const [route, methods] of Object.entries(paths)) {
    for (const [method, { parameters = [], responses = {} }] of Object.entries(methods)) {
      const segments = route
        .split("/")
        .filter(Boolean)
        .map((seg) => (seg.startsWith("{") ? `By${pascal(seg.slice(1, -1))}` : pascal(seg)));
      const name = pascal(method) + segments.join("");
      const body = parameters.find((p) => p.in === "body");
      const pathParams = parameters.filter((p) => p.in === "path");
      const query = parameters.filter((p) => p.in === "query");
      const ok = Object.keys(responses)
        .filter((code) => code.startsWith("2") && responses[code]?.schema)
        .sort()[0];

      const part = (suffix: string, schema: Schema | undefined) =>
        zod.push(`export const z${name}${suffix} = ${render(schema, emitted)};`, "");
      if (body) part("Body", body.schema);
      if (pathParams.length) part("Path", object(pathParams));
      if (query.length) part("Query", object(query));
      if (ok) part("Response", responses[ok]?.schema);

      const field = (key: string, present: boolean, required: boolean, suffix: string) =>
        present
          ? `  ${key}${required ? "" : "?"}: z.infer<typeof s.z${name}${suffix}>;`
          : `  ${key}?: never;`;
      types.push(
        `export type ${name}Data = {`,
        field("body", !!body, !!body?.required, "Body"),
        field("path", pathParams.length > 0, true, "Path"),
        field("query", query.length > 0, query.some((p) => p.required), "Query"),
        `  url: "${route}";`,
        "};",
        "",
      );
    }
  }

  mkdirSync(output, { recursive: true });
  writeFileSync(path.join(output, "zod.gen.ts"), zod.join("\n"));
  writeFileSync(path.join(output, "types.gen.ts"), types.join("\n"));
}
