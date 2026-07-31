import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { promisify } from "node:util";
import test from "node:test";
import { capture } from "./capture.js";
import type { CaptureConfiguration } from "./types.js";

const execFileAsync = promisify(execFile);

test(
  "captures a synthetic real-runtime map state in pinned Chromium",
  { skip: process.env.SOKU_BROWSER_E2E !== "1" },
  async () => {
    const root = await fs.mkdtemp(path.join(os.tmpdir(), "soku-browser-e2e-"));
    try {
      await fs.mkdir(path.join(root, "dist"), { recursive: true });
      await fs.mkdir(path.join(root, "docs", "manual", "fixtures"), { recursive: true });
      await fs.writeFile(
        path.join(root, "docs", "manual", "fixtures", "route.synthetic.json"),
        `${JSON.stringify({ state: "loaded" })}\n`,
        "utf8",
      );
      await fs.writeFile(
        path.join(root, "dist", "index.html"),
        `<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: Arial, sans-serif; margin: 0; padding: 24px; }
    .map { width: 360px; height: 420px; position: relative; overflow: hidden;
      background: linear-gradient(135deg, #dbeafe, #dcfce7); border: 2px solid #0f172a; }
    .marker { position: absolute; left: 160px; top: 160px; font-size: 32px; }
    .attribution { position: absolute; right: 4px; bottom: 4px; background: #fff;
      color: #111; padding: 3px 6px; font-size: 12px; }
  </style>
</head>
<body>
  <h1>配送ダッシュボード 😀</h1>
  <button type="button" data-testid="load">Load route</button>
  <p data-testid="state">idle</p>
  <section class="map map-ready" aria-label="Delivery map">
    <span class="marker">●</span>
    <span class="attribution">Deterministic test map © Project</span>
  </section>
  <script>
    document.querySelector('[data-testid="load"]').addEventListener('click', async () => {
      const response = await fetch('/api/dashboard');
      const payload = await response.json();
      document.querySelector('[data-testid="state"]').textContent = payload.state;
    });
  </script>
</body>
</html>`,
        "utf8",
      );
      const configuration: CaptureConfiguration = {
        schema_version: 1,
        execution: { mode: "local-manual", allow_hosts: [], request_budget: 0 },
        runtime: {
          adapter: "static-build",
          static_directory: "dist",
          serve: { origin: "loopback-http" },
        },
        browser: {
          engine: "chromium",
          viewport: { width: 414, height: 896 },
          device_scale_factor: 1,
          locale: "ja-JP",
          timezone: "Asia/Tokyo",
          fonts: ["Arial"],
        },
        backend: {
          mode: "http-route",
          routes: [
            {
              method: "GET",
              url: "/api/dashboard",
              status: 200,
              fixture: "docs/manual/fixtures/route.synthetic.json",
            },
          ],
        },
        map: {
          provider: "local-deterministic",
          source_relation: "original-application",
          execution_mode: "local-manual",
          map_load_budget: 1,
          request_budget: 0,
          readiness: { type: "deterministic", name: ".map-ready" },
          attribution: { required: true, locator: ".attribution" },
        },
        output: {
          directory: "docs/manual/captures",
          report: "docs/manual/capture-report.json",
          generated_index: "docs/manual/generated-index.md",
        },
        scenarios: [
          {
            id: "dashboard",
            route: "/",
            steps: [
              { action: "click", locator: { by: "test-id", value: "load" } },
              {
                action: "wait",
                locator: { by: "test-id", value: "state" },
                wait: { state: "text", value: "loaded" },
              },
              {
                action: "capture",
                id: "map-ready",
                capture: {
                  mode: "locator",
                  locator: { by: "css", value: ".map" },
                  padding: 48,
                  caption: "Original deterministic map with visible attribution.",
                },
              },
            ],
          },
        ],
      };
      await fs.writeFile(
        path.join(root, "docs", "manual", "capture.yml"),
        `${JSON.stringify(configuration, null, 2)}\n`,
        "utf8",
      );
      await execFileAsync("git", ["init", "--quiet"], { cwd: root });
      await execFileAsync("git", ["config", "user.name", "Soku Browser Test"], { cwd: root });
      await execFileAsync("git", ["config", "user.email", "browser-test@example.invalid"], { cwd: root });
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "test fixture"], { cwd: root });

      const report = await capture(root, "docs/manual/capture.yml", false);
      assert.equal(report.authenticity, "runtime-authentic-with-adapters");
      assert.ok(report.adapters.includes("backend:http-route"));
      assert.equal(report.map.attribution_visible, true);
      assert.equal(report.map.map_load_count, 1);
      assert.equal(report.map.request_count, 0);
      assert.equal(report.captures.length, 1);
      assert.equal(report.captures[0]?.clip?.x, 0);
      assert.equal(report.captures[0]?.clip?.width, 414);
      assert.equal(report.source.dirty, false);
      assert.equal(report.source.files[0]?.path, "dist/index.html");
      assert.equal(
        report.source.fixtures[0]?.path,
        "docs/manual/fixtures/route.synthetic.json",
      );
      const png = await fs.readFile(path.join(root, "docs", "manual", "captures", "map-ready.png"));
      assert.deepEqual([...png.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
      const durableReport = await fs.readFile(
        path.join(root, "docs", "manual", "capture-report.json"),
        "utf8",
      );
      assert.doesNotMatch(durableReport, /AIza|private|signature=/i);
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "capture baseline"], {
        cwd: root,
      });
      const replacementReport = await capture(root, "docs/manual/capture.yml", false);
      assert.equal(replacementReport.captures[0]?.capture_id, "map-ready");

      const desktopConfiguration = structuredClone(configuration);
      desktopConfiguration.browser.viewport = { width: 1280, height: 800 };
      desktopConfiguration.output.directory = "docs/manual/captures-desktop";
      desktopConfiguration.output.report = "docs/manual/capture-report-desktop.json";
      desktopConfiguration.output.generated_index = "docs/manual/generated-index-desktop.md";
      desktopConfiguration.scenarios[0]!.steps[2]!.id = "map-ready-desktop";
      await fs.writeFile(
        path.join(root, "docs", "manual", "capture.yml"),
        `${JSON.stringify(desktopConfiguration, null, 2)}\n`,
        "utf8",
      );
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "desktop fixture"], { cwd: root });
      const desktopReport = await capture(root, "docs/manual/capture.yml", false);
      assert.deepEqual(desktopReport.environment.viewport, { width: 1280, height: 800 });
      assert.equal(desktopReport.captures[0]?.capture_id, "map-ready-desktop");

      const googleAdapterConfiguration = structuredClone(desktopConfiguration);
      googleAdapterConfiguration.execution.allow_hosts = [
        "maps.googleapis.com",
        "maps.gstatic.com",
      ];
      googleAdapterConfiguration.execution.request_budget = 2;
      googleAdapterConfiguration.map = {
        provider: "google-maps-javascript",
        source_relation: "declared-adapter",
        api_key_env: "GOOGLE_MAPS_API_KEY",
        billing_owner: "documentation-owner",
        restriction_reviewed: true,
        execution_mode: "local-manual",
        map_load_budget: 1,
        request_budget: 2,
        readiness: { type: "deterministic", name: ".map-ready" },
        attribution: { required: true, locator: ".attribution" },
      };
      googleAdapterConfiguration.output.directory = "docs/manual/captures-google-adapter";
      googleAdapterConfiguration.output.report = "docs/manual/capture-report-google-adapter.json";
      googleAdapterConfiguration.output.generated_index =
        "docs/manual/generated-index-google-adapter.md";
      const googleCapture = googleAdapterConfiguration.scenarios[0]!.steps[2]!;
      googleCapture.id = "google-test-adapter";
      if (googleCapture.capture === undefined) throw new Error("capture fixture missing");
      googleCapture.capture.caption =
        "Google Maps JavaScript test adapter with visible synthetic attribution.";
      await fs.writeFile(
        path.join(root, "docs", "manual", "capture.yml"),
        `${JSON.stringify(googleAdapterConfiguration, null, 2)}\n`,
        "utf8",
      );
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "google adapter fixture"], {
        cwd: root,
      });
      process.env.GOOGLE_MAPS_API_KEY = "synthetic-test-value";
      const googleReport = await capture(root, "docs/manual/capture.yml", false);
      delete process.env.GOOGLE_MAPS_API_KEY;
      assert.equal(googleReport.authenticity, "runtime-authentic-with-adapters");
      assert.equal(googleReport.map.api_key_env, "GOOGLE_MAPS_API_KEY");
      assert.ok(googleReport.adapters.includes("map:google-maps-javascript"));
      assert.doesNotMatch(JSON.stringify(googleReport), /synthetic-test-value/);

      const cropConfiguration = structuredClone(desktopConfiguration);
      cropConfiguration.output.directory = "docs/manual/captures-crop";
      cropConfiguration.output.report = "docs/manual/capture-report-crop.json";
      cropConfiguration.output.generated_index = "docs/manual/generated-index-crop.md";
      const cropStep = cropConfiguration.scenarios[0]!.steps[2]!;
      cropStep.id = "map-cropped";
      cropStep.capture = {
        mode: "region",
        region: { x: 0, y: 0, width: 100, height: 100 },
        caption: "Map region that must be refused because attribution is outside the clip.",
      };
      await fs.writeFile(
        path.join(root, "docs", "manual", "capture.yml"),
        `${JSON.stringify(cropConfiguration, null, 2)}\n`,
        "utf8",
      );
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "crop refusal fixture"], {
        cwd: root,
      });
      await assert.rejects(
        () => capture(root, "docs/manual/capture.yml", false),
        /crop or obscure required map attribution/,
      );

      const missingFontConfiguration = structuredClone(desktopConfiguration);
      missingFontConfiguration.browser.fonts = ["Definitely Missing Soku Font"];
      missingFontConfiguration.output.report = "docs/manual/capture-report-font-failure.json";
      await fs.writeFile(
        path.join(root, "docs", "manual", "capture.yml"),
        `${JSON.stringify(missingFontConfiguration, null, 2)}\n`,
        "utf8",
      );
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "font failure fixture"], { cwd: root });
      await assert.rejects(
        () => capture(root, "docs/manual/capture.yml", false),
        /Japanese\/emoji font readiness check failed/,
      );
    } finally {
      await fs.rm(root, { recursive: true, force: true });
    }
  },
);

test(
  "assembles GAS fragments and captures disclosed async bridge and dialog states",
  { skip: process.env.SOKU_BROWSER_E2E !== "1" },
  async () => {
    const root = await fs.mkdtemp(path.join(os.tmpdir(), "soku-gas-e2e-"));
    try {
      await fs.mkdir(path.join(root, "gas"), { recursive: true });
      await fs.mkdir(path.join(root, "docs", "manual", "fixtures"), { recursive: true });
      await fs.writeFile(
        path.join(root, "gas", "Style.html"),
        `<style>
          body { font-family: Arial, sans-serif; padding: 24px; }
          [data-soku-dialog-overlay] button { min-width: 120px; }
        </style>`,
        "utf8",
      );
      await fs.writeFile(
        path.join(root, "gas", "Index.html"),
        `<h1>GAS 配送管理 😀</h1>
        <button id="load" type="button">Load</button>
        <button id="fail" type="button">Fail</button>
        <p id="state">idle</p>`,
        "utf8",
      );
      await fs.writeFile(
        path.join(root, "gas", "Script.html"),
        `<script>
          document.querySelector('#load').addEventListener('click', () => {
            google.script.run
              .withUserObject({source: 'manual'})
              .withSuccessHandler((value, user) => {
                document.querySelector('#state').textContent = 'loaded:' + value.count + ':' + user.source;
              })
              .withFailureHandler((error) => {
                document.querySelector('#state').textContent = error.message;
              })
              .loadDashboard();
          });
          document.querySelector('#fail').addEventListener('click', () => {
            google.script.run
              .withFailureHandler((error) => {
                document.querySelector('#state').textContent = error.message;
              })
              .failDashboard();
          });
        </script>`,
        "utf8",
      );
      await fs.writeFile(
        path.join(root, "docs", "manual", "fixtures", "dashboard.synthetic.json"),
        `${JSON.stringify({
          state: { count: 1 },
          methods: {
            loadDashboard: { result: "$state", mutate: { count: 2 } },
            failDashboard: { error: "synthetic failure" },
          },
        })}\n`,
        "utf8",
      );
      const configuration: CaptureConfiguration = {
        schema_version: 1,
        execution: { mode: "local-manual", allow_hosts: [], request_budget: 0 },
        runtime: {
          adapter: "gas-html-service",
          source_fragments: ["gas/Style.html", "gas/Index.html", "gas/Script.html"],
          serve: { origin: "loopback-http" },
        },
        browser: {
          engine: "chromium",
          viewport: { width: 414, height: 896 },
          device_scale_factor: 1,
          locale: "ja-JP",
          timezone: "Asia/Tokyo",
          fonts: ["Arial"],
        },
        backend: {
          mode: "browser-bridge",
          adapter: "google-script-run",
          fixture: "docs/manual/fixtures/dashboard.synthetic.json",
        },
        map: {
          provider: "none",
          source_relation: "not-applicable",
          execution_mode: "local-manual",
          map_load_budget: 0,
          request_budget: 0,
          readiness: { type: "deterministic", name: "not-applicable" },
          attribution: { required: false },
        },
        output: {
          directory: "docs/manual/captures",
          report: "docs/manual/capture-report.json",
          generated_index: "docs/manual/generated-index.md",
        },
        scenarios: [
          {
            id: "gas-dashboard",
            route: "/",
            steps: [
              { action: "click", locator: { by: "css", value: "#load" } },
              {
                action: "wait",
                locator: { by: "css", value: "#state" },
                wait: { state: "text", value: "loaded:2:manual" },
              },
              { action: "click", locator: { by: "css", value: "#fail" } },
              {
                action: "wait",
                locator: { by: "css", value: "#state" },
                wait: { state: "text", value: "synthetic failure" },
              },
              {
                action: "dialog",
                dialog: {
                  mode: "documentation-overlay",
                  action: "accept",
                  message: "Continue with the synthetic update?",
                },
              },
              {
                action: "capture",
                id: "gas-confirmation",
                capture: {
                  mode: "locator",
                  locator: { by: "css", value: "[data-soku-dialog-overlay]" },
                  caption: "Documentation overlay adapter for the GAS confirmation dialog.",
                },
              },
            ],
          },
        ],
      };
      await fs.writeFile(
        path.join(root, "docs", "manual", "capture.yml"),
        `${JSON.stringify(configuration, null, 2)}\n`,
        "utf8",
      );
      await execFileAsync("git", ["init", "--quiet"], { cwd: root });
      await execFileAsync("git", ["config", "user.name", "Soku GAS Test"], { cwd: root });
      await execFileAsync("git", ["config", "user.email", "gas-test@example.invalid"], { cwd: root });
      await execFileAsync("git", ["add", "."], { cwd: root });
      await execFileAsync("git", ["commit", "--quiet", "-m", "test fixture"], { cwd: root });

      const report = await capture(root, "docs/manual/capture.yml", false);
      assert.equal(report.authenticity, "runtime-authentic-with-adapters");
      assert.deepEqual(report.adapters, [
        "backend:browser-bridge",
        "dialog:documentation-overlay",
        "runtime:gas-html-service",
      ]);
      assert.equal(report.captures[0]?.dialog_adapter, "documentation-overlay");
      assert.equal(report.source.files.length, 3);
      assert.equal(report.source.fixtures[0]?.path, "docs/manual/fixtures/dashboard.synthetic.json");
    } finally {
      await fs.rm(root, { recursive: true, force: true });
    }
  },
);
