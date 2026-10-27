import { readFileSync, rmSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parse } from "yaml";
import { emit } from "./emit-zod.mts";
import { GoIndex, applyPresence, isNode } from "./go-presence.mts";
import { settle, supplement } from "./spec-supplement.mts";

const here = path.dirname(fileURLToPath(import.meta.url));
const webRoot = path.resolve(here, "..");
const repoRoot = path.resolve(webRoot, "..");

type ModuleSpec = {
  name: string;
  input: string;
  output: string;
};

const modules: ModuleSpec[] = [
  {
    name: "core",
    input: path.join(repoRoot, "api", "core.openapi.yaml"),
    output: path.join(webRoot, "src", "generated", "core"),
  },
  {
    name: "amie",
    input: path.join(
      repoRoot,
      "connectors",
      "ACCESS",
      "AMIE-Processor",
      "api",
      "amie.openapi.yaml",
    ),
    output: path.join(webRoot, "src", "generated", "amie"),
  },
  {
    name: "analytics",
    input: path.join(repoRoot, "connectors", "Analytics", "api", "analytics.openapi.yaml"),
    output: path.join(webRoot, "src", "generated", "analytics"),
  },
];

async function main() {
  const go = new GoIndex(repoRoot);
  for (const mod of modules) {
    const spec: unknown = parse(readFileSync(mod.input, "utf8"));
    if (!isNode(spec)) throw new Error(`gen-api: ${mod.input} is not an OpenAPI document`);
    supplement(mod.name, spec, go, repoRoot);
    const report = applyPresence(spec, go);
    settle(mod.name, spec);
    console.log(`gen-api: ${mod.name} presence ${JSON.stringify(report)}`);
    rmSync(mod.output, { recursive: true, force: true });
    emit(spec, mod.output);
    console.log(`gen-api: emitted ${mod.name} → ${path.relative(webRoot, mod.output)}`);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
