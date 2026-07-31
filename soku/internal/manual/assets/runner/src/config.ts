import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { Ajv2020, type ErrorObject } from "ajv/dist/2020.js";
import addFormatsModule from "ajv-formats";
import { parseDocument } from "yaml";
import { validateProviderConfiguration } from "./providers.js";
import type { CaptureConfiguration, CaptureReport } from "./types.js";

const credentialLiteral =
  /(AIza[0-9A-Za-z_-]{20,}|(?:api[_-]?key|token|secret|signature|password)\s*:\s*[^$\s][^\r\n#]*)/i;

export function portablePath(value: string): boolean {
  if (
    value.length === 0 ||
    value.startsWith("/") ||
    value.includes("\\") ||
    /^[A-Za-z]:/.test(value)
  ) {
    return false;
  }
  const normalized = path.posix.normalize(value);
  return (
    normalized === value &&
    normalized !== "." &&
    !normalized.split("/").some((part) => part === ".." || /^\.git$/i.test(part) || /^\.soku$/i.test(part))
  );
}

export async function loadConfiguration(
  root: string,
  configPath: string,
): Promise<CaptureConfiguration> {
  if (!portablePath(configPath)) {
    throw new Error("configuration path is not a portable repository-relative path");
  }
  const raw = await fs.readFile(path.join(root, ...configPath.split("/")), "utf8");
  if (credentialLiteral.test(raw)) {
    throw new Error("configuration contains a credential-like literal");
  }
  const document = parseDocument(raw, {
    uniqueKeys: true,
    strict: true,
  });
  if (document.errors.length > 0) {
    throw new Error(`invalid YAML: ${document.errors[0]?.message ?? "unknown error"}`);
  }
  const value: unknown = document.toJS();
  const schema = await readSchema("manual-capture-v1.schema.json");
  validateAgainstSchema(schema, value, "capture configuration");
  const config = value as CaptureConfiguration;
  validateSemantics(config);
  return config;
}

export async function validateCaptureReport(report: CaptureReport): Promise<void> {
  const schema = await readSchema("manual-capture-report-v1.schema.json");
  validateAgainstSchema(schema, report, "capture report");
}

async function readSchema(name: string): Promise<object> {
  const runnerDirectory = path.dirname(fileURLToPath(import.meta.url));
  const installedSchema = path.resolve(
    runnerDirectory, "..", "..", "..", "docs", "manual", "schema", name,
  );
  const developmentSchema = path.resolve(
    runnerDirectory, "..", "..", "schemas", name,
  );
  const schemaPath = await fs.access(installedSchema).then(
    () => installedSchema,
    () => developmentSchema,
  );
  return JSON.parse(await fs.readFile(schemaPath, "utf8")) as object;
}

function validateAgainstSchema(
  schema: object,
  value: unknown,
  contractName: string,
): void {
  const ajv = new Ajv2020({ allErrors: true, strict: true, strictRequired: false });
  type FormatsPlugin = (instance: Ajv2020) => Ajv2020;
  const addFormats =
    (addFormatsModule as unknown as { default?: FormatsPlugin }).default ??
    (addFormatsModule as unknown as FormatsPlugin);
  addFormats(ajv);
  const validate = ajv.compile(schema);
  if (!validate(value)) {
    const message = validate.errors
      ?.map((error: ErrorObject) => `${error.instancePath || "/"} ${error.message ?? "is invalid"}`)
      .join("; ");
    throw new Error(`${contractName} schema validation failed: ${message}`);
  }
}

function validateSemantics(config: CaptureConfiguration): void {
  if (
    [...config.execution.allow_hosts].sort().join("\0") !==
    config.execution.allow_hosts.join("\0")
  ) {
    throw new Error("execution.allow_hosts must be sorted");
  }
  const captureIds = new Set<string>();
  for (const scenario of config.scenarios) {
    let overlayPending = false;
    for (const step of scenario.steps) {
      if (
        step.action === "wait" &&
        step.wait?.state === "attribute" &&
        !/^[^=\s]+=/.test(step.wait.value ?? "")
      ) {
        throw new Error("attribute wait value must use name=value");
      }
      if (step.action === "dialog" && step.dialog?.mode === "documentation-overlay") {
        overlayPending = true;
      }
      if (step.action === "capture") {
        if (step.id === undefined || captureIds.has(step.id)) {
          throw new Error("capture ids must be present and unique");
        }
        captureIds.add(step.id);
        if (
          config.map.source_relation === "declared-adapter" &&
          !step.capture?.caption.toLowerCase().includes("adapter")
        ) {
          throw new Error("map provider substitution requires adapter disclosure");
        }
        if (
          overlayPending &&
          (!step.capture?.caption.toLowerCase().includes("documentation") ||
            !step.capture.caption.toLowerCase().includes("overlay"))
        ) {
          throw new Error("documentation-overlay capture caption must disclose the adapter");
        }
        overlayPending = false;
      }
    }
  }
  const paths = [
    config.output.directory,
    config.output.report,
    config.output.generated_index,
    config.backend.fixture,
    config.backend.har,
    config.runtime.static_directory,
    config.pdf?.path,
    config.pdf?.raster_output,
    ...(config.backend.routes ?? []).map((route) => route.fixture),
    ...(config.runtime.source_files ?? []),
    ...(config.runtime.source_fragments ?? []),
    ...(config.hooks ?? []).map((hook) => hook.path),
  ].filter((value): value is string => value !== undefined);
  if (paths.some((value) => !portablePath(value))) {
    throw new Error("configuration contains an unsafe project path");
  }
  if (
    config.backend.mode === "har-replay" &&
    !config.backend.har?.toLowerCase().endsWith(".sanitized.har")
  ) {
    throw new Error("HAR replay requires a *.sanitized.har path");
  }
  if (
    [config.backend.fixture, ...(config.backend.routes ?? []).map((route) => route.fixture)]
      .filter((value): value is string => value !== undefined)
      .some((value) => !value.includes(".synthetic."))
  ) {
    throw new Error("backend fixture filenames must disclose synthetic data");
  }
  validateProviderConfiguration(config);
}
