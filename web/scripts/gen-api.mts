import path from "node:path";
import { fileURLToPath } from "node:url";
import { createClient } from "@hey-api/openapi-ts";

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
  {
    name: "temp-account",
    input: path.join(repoRoot, "connectors", "TempAccount", "api", "temp-account.openapi.yaml"),
    output: path.join(webRoot, "src", "generated", "temp-account"),
  },
];

async function main() {
  for (const mod of modules) {
    await createClient({
      input: mod.input,
      output: { path: mod.output, postProcess: [] },
      plugins: [
        {
          name: "@hey-api/client-fetch",
          baseUrl: "/api/v1",
          runtimeConfigPath: "@/shared/api/client",
          throwOnError: true,
        },
        { name: "@hey-api/typescript" },
        { name: "@hey-api/sdk", responseStyle: "data" },
        {
          name: "zod",
          definitions: { types: { infer: true } },
        },
      ],
    });
    console.log(`gen-api: emitted ${mod.name} → ${path.relative(webRoot, mod.output)}`);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
