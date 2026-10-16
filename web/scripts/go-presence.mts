/**
 * Swag marks almost no property required, so this pass restates each response schema's presence and
 * nullability from the Go struct swag generated it from (`--useStructName` names it by bare struct name),
 * read as data at gen time. encoding/json decides: omitempty or omitzero → optional; a pointer, slice,
 * map, RawMessage or interface without it → required and nullable; anything else → required. A schema
 * maps only when exactly one rule set comes from structs whose JSON fields equal its properties;
 * request-only schemas are left alone, since decoding tolerates absence.
 */
import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

type Prop = { required: boolean; nullable: boolean };
type Decl = { name: string; pkg: string; dir: string } & (
  | { kind: "struct"; body: string }
  | { kind: "named"; underlying: string }
);
type Marshal = { custom: boolean; nonNil: Set<string> };
// biome-ignore lint/suspicious/noExplicitAny: an OpenAPI document is untyped YAML
type Json = any;

const skipDirs = new Set(["web", "node_modules", "vendor", ".git", "dev-ops"]);

function goFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) =>
    e.isDirectory()
      ? skipDirs.has(e.name)
        ? []
        : goFiles(path.join(dir, e.name))
      : e.name.endsWith(".go") && !e.name.endsWith("_test.go")
        ? [path.join(dir, e.name)]
        : [],
  );
}

function closing(src: string, open: number, [o, c]: string): number {
  let depth = 0;
  for (let i = open; i < src.length; i++) {
    if (src[i] === "`") i = src.indexOf("`", i + 1);
    else if (src[i] === o) depth++;
    else if (src[i] === c && --depth === 0) return i;
  }
  return src.length;
}

export class GoIndex {
  private decls = new Map<string, Decl[]>();
  private marshal = new Map<string, Marshal>();

  constructor(root: string) {
    for (const file of goFiles(root)) {
      const src = readFileSync(file, "utf8")
        .replace(/\/\*[\s\S]*?\*\//g, "")
        .replace(/\/\/.*$/gm, "");
      const pkg = /^package\s+(\w+)/m.exec(src)?.[1] ?? "";
      const dir = path.dirname(file);
      const add = (decl: Decl) =>
        this.decls.set(decl.name, [...(this.decls.get(decl.name) ?? []), decl]);
      const declare = (text: string, offset: number, group: boolean) => {
        const re = group
          ? /^\s*(\w+)\s+(?:=\s*)?(struct\s*\{|[^\n]+)/gm
          : /\btype\s+(\w+)\s+(?:=\s*)?(struct\s*\{|[^\n]+)/g;
        for (let m = re.exec(text); m; m = re.exec(text)) {
          const [whole, name = "", underlying = ""] = m;
          if (!underlying.startsWith("struct")) {
            add({ name, pkg, dir, kind: "named", underlying: underlying.trim() });
          } else {
            const open = offset + m.index + whole.length - 1;
            const end = closing(src, open, "{}");
            add({ name, pkg, dir, kind: "struct", body: src.slice(open + 1, end) });
            if (group) re.lastIndex = end - offset;
          }
        }
      };
      declare(src, 0, false);
      for (let m = /\btype\s*\(/g, g = m.exec(src); g; g = m.exec(src)) {
        const open = g.index + g[0].length - 1;
        declare(src.slice(open + 1, closing(src, open, "()")), open + 1, true);
      }
      const fn = /func\s*\(\s*\w+\s+\*?(\w+)\s*\)\s*MarshalJSON\s*\(\s*\)[^{]*\{/g;
      for (let m = fn.exec(src); m; m = fn.exec(src)) {
        const name = m[1] ?? "";
        const open = m.index + m[0].length - 1;
        const body = src.slice(open, closing(src, open, "{}"));
        const aliased = new RegExp(`\\btype\\s+\\w+\\s+${name}\\b`).test(body);
        const nonNil = new Set(
          [...body.matchAll(/\.(\w+)\s*=\s*(?:\[\]|map\[|make\()/g)].map((a) => a[1] ?? ""),
        );
        this.marshal.set(`${dir}:${name}`, { custom: !aliased, nonNil });
      }
    }
  }

  private resolve(ref: string, from: { dir: string }): Decl[] {
    const [qual, name] = ref.includes(".") ? ref.split(".") : [undefined, ref];
    return (this.decls.get(name ?? "") ?? []).filter((d) =>
      qual ? d.pkg === qual : d.dir === from.dir,
    );
  }

  private props(decl: Decl, seen = new Set<Decl>()): Map<string, Prop> | undefined {
    if (seen.has(decl)) return undefined;
    seen.add(decl);
    if (decl.kind === "named") {
      const [target] = this.resolve(decl.underlying, decl);
      return target && this.props(target, seen);
    }
    const marshal = this.marshal.get(`${decl.dir}:${decl.name}`);
    if (marshal?.custom) return undefined;
    const out = new Map<string, Prop>();
    const body = decl.body.replace(/\{[^{}`]*\}/g, "{}");
    for (const line of body
      .split(/[\n;]/)
      .map((l) => l.trim())
      .filter(Boolean)) {
      const m = /^(?:(\w+(?:\s*,\s*\w+)*)\s+)?([^`]+?)\s*(`[^`]*`)?$/.exec(line);
      if (!m) return undefined;
      const [, names, type = "", tag = ""] = m;
      const [jsonName = "", ...opts] = /json:"([^"]*)"/.exec(tag)?.[1]?.split(",") ?? [];
      if (jsonName === "-" && opts.length === 0) continue;
      const optional = opts.includes("omitempty") || opts.includes("omitzero");
      if (!names && !jsonName) {
        const [embedded] = this.resolve(type.replace(/^\*/, ""), decl);
        const inner = embedded && this.props(embedded, seen);
        if (!inner) return undefined;
        for (const [k, p] of inner)
          out.set(k, type.startsWith("*") ? { ...p, required: false } : p);
        continue;
      }
      for (const field of names?.split(/\s*,\s*/) ?? [type.replace(/^\*?(\w+\.)?/, "")]) {
        if (!/^[A-Z]/.test(field)) continue;
        const nullable = this.nullable(type, decl) && !marshal?.nonNil.has(field);
        out.set(jsonName || field, { required: !optional, nullable: !optional && nullable });
      }
    }
    return out;
  }

  // A named type answers by its underlying type unless it marshals itself.
  private nullable(type: string, from: Decl): boolean {
    if (/^(\*|\[\]|map\[|interface\s*\{|any$|json\.RawMessage$)/.test(type)) return true;
    const [named] = this.resolve(type, from);
    if (named?.kind !== "named" || this.marshal.has(`${named.dir}:${named.name}`)) return false;
    return this.nullable(named.underlying, named);
  }

  /** The presence rules for a schema, or undefined when no struct (or more than one) matches it. */
  rules(name: string, properties: string[]): Map<string, Prop> | undefined {
    const want = [...properties].sort().join();
    const found = new Set<string>();
    let rules: Map<string, Prop> | undefined;
    for (const decl of this.decls.get(name) ?? []) {
      const p = this.props(decl);
      if (!p || [...p.keys()].sort().join() !== want) continue;
      found.add(JSON.stringify([...p]));
      rules = p;
    }
    return found.size === 1 ? rules : undefined;
  }
}

function refs(node: Json, out = new Set<string>()): Set<string> {
  if (node && typeof node === "object") {
    if (typeof node.$ref === "string") out.add(node.$ref.replace("#/definitions/", ""));
    for (const v of Object.values(node)) refs(v, out);
  }
  return out;
}

function closure(roots: Set<string>, defs: Record<string, Json>): Set<string> {
  const out = new Set<string>();
  const visit = (n: string) => {
    if (out.has(n) || !defs[n]) return;
    out.add(n);
    for (const r of refs(defs[n])) visit(r);
  };
  for (const r of roots) visit(r);
  return out;
}

function retarget(node: Json, names: Set<string>): void {
  if (!node || typeof node !== "object") return;
  const name = typeof node.$ref === "string" && node.$ref.replace("#/definitions/", "");
  if (name && names.has(name)) node.$ref = `#/definitions/${name}Input`;
  for (const v of Object.values(node)) retarget(v, names);
}

export type Report = { required: number; nullable: number; schemas: number; unmapped: string[] };

/**
 * Rewrites the response schemas of a Swagger 2 document in place. A schema that request bodies share
 * keeps its loose form under `<Name>Input`, which those bodies reference instead.
 */
export function applyPresence(spec: Json, index: GoIndex): Report {
  const defs: Record<string, Json> = spec.definitions ?? {};
  const ops = Object.entries(spec.paths ?? {}).flatMap(([route, item]) =>
    Object.entries(item as Json).map(([method, op]) => ({ ...(op as Json), route, method })),
  ) as Json[];
  const bodies = ops.flatMap((o) => o.parameters ?? []).filter((p: Json) => p.in === "body");
  const responses = closure(refs(ops.map((o) => o.responses)), defs);
  const shared = new Set(
    [...closure(refs(bodies), defs)].filter((n) => responses.has(n) && defs[n].properties),
  );
  for (const name of shared) {
    defs[`${name}Input`] = structuredClone(defs[name]);
    retarget(defs[`${name}Input`], shared);
  }
  retarget(bodies, shared);
  // An inline response object comes from a handler's map literal, not a struct; the supplement
  // states the ones the portal reads.
  const inline = ops.flatMap((o) =>
    Object.entries(o.responses ?? {})
      .filter(([code, r]) => code.startsWith("2") && (r as Json).schema?.properties)
      .filter(([, r]) => !(r as Json).schema.required)
      .map(() => `${o.method.toUpperCase()} ${o.route}`),
  );
  const report: Report = { required: 0, nullable: 0, schemas: 0, unmapped: inline };
  for (const name of responses) {
    const props: Record<string, Json> | undefined = defs[name]?.properties;
    if (!props) continue;
    const rules = index.rules(name, Object.keys(props));
    if (!rules) {
      report.unmapped.push(name);
      continue;
    }
    report.schemas++;
    defs[name].required = [...rules].filter(([, p]) => p.required).map(([k]) => k);
    report.required += defs[name].required.length;
    for (const [key, p] of rules) {
      if (!p.nullable) continue;
      const { $ref, ...rest } = props[key];
      props[key] = { ...rest, ...($ref ? { allOf: [{ $ref }] } : {}), "x-nullable": true };
      report.nullable++;
    }
  }
  return report;
}
