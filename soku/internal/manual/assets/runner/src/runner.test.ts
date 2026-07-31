import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import { loadConfiguration, validateCaptureReport } from "./config.js";
import { gasBridgeSource } from "./gas-bridge.js";
import {
  replaceGeneratedFiles,
  reportIntegritySHA256,
  sha256File,
  stableAssetName,
} from "./output.js";
import type { CaptureReport } from "./types.js";
import { assertRedacted, redactText, redactURL } from "./redaction.js";

test("redacts query credentials and literals", () => {
  const redacted = redactURL("https://maps.example.test/view?key=AIza012345678901234567890123&x=1");
  assert.match(redacted, /key=%5BREDACTED%5D/);
  assert.doesNotMatch(redacted, /AIza/);
  assert.equal(redactText("token=private"), "token=[REDACTED]");
  assert.throws(() => assertRedacted({ url: "https://x.test/?signature=private" }));
});

test("stable asset names reject path-like ids", () => {
  assert.equal(stableAssetName("dashboard-ready"), "dashboard-ready.png");
  assert.throws(() => stableAssetName("../escape"));
});

test("published example passes strict schema and semantic validation", async () => {
  const runnerDirectory = path.dirname(fileURLToPath(import.meta.url));
  const root = path.resolve(runnerDirectory, "..", "..");
  const examples = (await fs.readdir(path.join(root, "examples")))
    .filter((name) => name.endsWith(".yml"))
    .sort();
  assert.deepEqual(examples, [
    "capture.example.yml",
    "capture.gas-leaflet.example.yml",
    "capture.google-maps.example.yml",
  ]);
  const configurations = await Promise.all(
    examples.map((name) => loadConfiguration(root, `examples/${name}`)),
  );
  assert.deepEqual(
    configurations.map((config) => config.map.provider).sort(),
    ["google-maps-javascript", "leaflet-osm", "none"],
  );
});

test("GAS bridge preserves chainable asynchronous handlers", () => {
  const source = gasBridgeSource({
    methods: {
      loadDashboard: { result: "$state", mutate: { count: 2 } },
      failDashboard: { error: "synthetic failure" },
    },
    state: { count: 1 },
  });
  assert.match(source, /withSuccessHandler/);
  assert.match(source, /withFailureHandler/);
  assert.match(source, /withUserObject/);
  assert.match(source, /queueMicrotask/);
  assert.doesNotMatch(source, /<script/i);
});

test("output replacement refuses unowned paths and preserves modified generated files", async () => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "manual-output-"));
  const staging = await fs.mkdtemp(path.join(os.tmpdir(), "manual-stage-"));
  try {
    await fs.mkdir(path.join(root, "docs"), { recursive: true });
    await fs.writeFile(path.join(root, "docs", "capture.png"), "user");
    await fs.mkdir(path.join(staging, "docs"), { recursive: true });
    await fs.writeFile(path.join(staging, "docs", "capture.png"), "generated");
    const next = [{ path: "docs/capture.png", sha256: await sha256File(path.join(staging, "docs", "capture.png")) }];
    await assert.rejects(() => replaceGeneratedFiles(root, staging, next, "docs/report.json"), /unowned/);
  } finally {
    await fs.rm(root, { recursive: true, force: true });
    await fs.rm(staging, { recursive: true, force: true });
  }
});

test("report schema and integrity detect manual report edits", async () => {
  const report = {
    schema_version: 1,
    report_integrity_sha256: "0".repeat(64),
    authenticity: "runtime-authentic",
    source: { commit: "0".repeat(40), dirty: false, files: [], fixtures: [], hooks: [] },
    environment: {
      browser: "chromium",
      browser_version: "test",
      viewport: { width: 414, height: 896 },
      locale: "ja-JP",
      timezone: "Asia/Tokyo",
      device_scale_factor: 1,
      fonts_ready: true,
    },
    adapters: [],
    map: {
      provider: "none",
      source_relation: "not-applicable",
      execution_mode: "local-manual",
      billing_owner_declared: false,
      restriction_reviewed: false,
      readiness: "hook:not-applicable",
      attribution_visible: true,
      map_load_count: 0,
      request_count: 0,
    },
    egress: [],
    captures: [],
    generated_files: [],
  } satisfies CaptureReport;
  report.report_integrity_sha256 = reportIntegritySHA256(report);
  await validateCaptureReport(report);
  assert.equal(report.report_integrity_sha256, reportIntegritySHA256(report));
  report.environment.browser_version = "manually edited";
  assert.notEqual(report.report_integrity_sha256, reportIntegritySHA256(report));
  await assert.rejects(
    () => validateCaptureReport({ ...report, unexpected: true } as CaptureReport),
    /capture report schema validation failed/,
  );

  const root = await fs.mkdtemp(path.join(os.tmpdir(), "manual-report-"));
  const staging = await fs.mkdtemp(path.join(os.tmpdir(), "manual-report-stage-"));
  try {
    await fs.mkdir(path.join(root, "docs"), { recursive: true });
    await fs.writeFile(
      path.join(root, "docs", "report.json"),
      `${JSON.stringify(report, null, 2)}\n`,
      "utf8",
    );
    await assert.rejects(
      () => replaceGeneratedFiles(root, staging, [], "docs/report.json"),
      /generated output was modified or removed/,
    );
  } finally {
    await fs.rm(root, { recursive: true, force: true });
    await fs.rm(staging, { recursive: true, force: true });
  }
});
