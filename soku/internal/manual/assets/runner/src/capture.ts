import crypto from "node:crypto";
import fs from "node:fs/promises";
import path from "node:path";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { chromium, type BrowserContext, type Locator, type Page } from "playwright";
import { loadConfiguration, validateCaptureReport } from "./config.js";
import { runDialog } from "./dialogs.js";
import { assertRedacted, redactText, redactURL } from "./redaction.js";
import {
  replaceGeneratedFiles,
  reportIntegritySHA256,
  sha256File,
  stableAssetName,
} from "./output.js";
import {
  assertAttributionInClip,
  verifyAttribution,
  waitForMap,
} from "./providers.js";
import { startRuntime } from "./server.js";
import type {
  CaptureAsset,
  CaptureConfiguration,
  CaptureReport,
  HashRecord,
  LocatorConfig,
  StepConfig,
} from "./types.js";

const execFileAsync = promisify(execFile);

export async function probe(root: string, configPath: string): Promise<string> {
  const config = await loadConfiguration(root, configPath);
  if (config.execution.mode !== "local-manual") throw new Error("probe is local/manual only");
  if (
    config.map.provider === "google-maps-javascript" &&
    !process.env[config.map.api_key_env ?? ""]
  ) {
    throw new Error(`${config.map.api_key_env ?? "map key environment"} is absent`);
  }
  const browser = await chromium.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setContent("<p>probe</p>");
    await page.getByText("probe").waitFor({ state: "visible" });
    return `Chromium ${browser.version()} probe passed; no credential values were emitted.`;
  } finally {
    await browser.close();
  }
}

export async function capture(
  root: string,
  configPath: string,
  allowDirty: boolean,
): Promise<CaptureReport> {
  const config = await loadConfiguration(root, configPath);
  if (
    config.map.provider === "google-maps-javascript" &&
    !process.env[config.map.api_key_env ?? ""]
  ) {
    throw new Error(`${config.map.api_key_env ?? "map key environment"} is absent`);
  }
  const git = await gitIdentity(root);
  if (git.dirty && !allowDirty) {
    throw new Error("worktree is dirty; rerun with --allow-dirty to record that state");
  }
  const stagingRoot = await fs.mkdtemp(path.join(root, ".soku-manual-capture-"));
  const runtime = await startRuntime(root, config);
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: config.browser.viewport,
    deviceScaleFactor: config.browser.device_scale_factor,
    locale: config.browser.locale,
    timezoneId: config.browser.timezone,
  });
  const egress: Array<{ method: string; url: string; host: string }> = [];
  let externalRequests = 0;
  let mapLoads = 0;
  let attributionVisible = config.map.provider === "none";
  const assets: CaptureAsset[] = [];
  try {
    await installNetworkBoundary(context, config, runtime.baseURL, egress, () => {
      externalRequests += 1;
    });
    if (config.backend.mode === "har-replay" && config.backend.har !== undefined) {
      await context.routeFromHAR(path.join(root, ...config.backend.har.split("/")), {
        notFound: "abort",
        update: false,
      });
    }
    const page = await context.newPage();
    await installHTTPMocks(page, root, config);
    const consoleErrors: string[] = [];
    page.on("console", (message) => {
      if (message.type() === "error") consoleErrors.push(redactText(message.text()));
    });
    for (const scenario of config.scenarios) {
      await page.goto(new URL(scenario.route, runtime.baseURL).toString(), {
        waitUntil: "domcontentloaded",
      });
      await waitForFonts(page, config.browser.fonts);
      if (config.map.provider !== "none") {
        await waitForMap(page, config);
        mapLoads += 1;
        attributionVisible = await verifyAttribution(page, config);
      }
      for (const [stepIndex, step] of scenario.steps.entries()) {
        const asset = await runStep(page, step, scenario.id, stepIndex, stagingRoot, config);
        if (asset !== undefined) assets.push(asset);
      }
    }
    if (consoleErrors.some((value) => /(google maps|invalidkey|billing|referer)/i.test(value))) {
      throw new Error(`map provider console error: ${consoleErrors.join("; ")}`);
    }
    if (
      mapLoads > config.map.map_load_budget ||
      externalRequests > config.execution.request_budget ||
      externalRequests > config.map.request_budget
    ) {
      throw new Error("declared map-load or external-request budget was exceeded");
    }
    const sourceFiles = unique([
      ...runtime.sourceFiles,
      ...(config.runtime.source_files ?? []),
      ...(config.runtime.source_fragments ?? []),
    ]);
    const fixtureFiles = unique([
      config.backend.fixture,
      ...(config.backend.routes ?? []).map((route) => route.fixture),
      config.backend.har,
    ].filter((value): value is string => value !== undefined));
    const hookFiles = unique((config.hooks ?? []).map((hook) => hook.path));
    const pdfReview = await reviewPDF(root, stagingRoot, config);
    const generatedIndex = renderIndex(config, assets);
    const indexStaging = stagedPath(stagingRoot, config.output.generated_index);
    await fs.mkdir(path.dirname(indexStaging), { recursive: true });
    await fs.writeFile(indexStaging, generatedIndex, "utf8");
    const generated: HashRecord[] = await Promise.all([
      ...assets.map(async (asset) => ({ path: asset.path, sha256: asset.sha256 })),
      {
        path: config.output.generated_index,
        sha256: await sha256File(indexStaging),
      },
      ...(pdfReview?.raster_pages ?? []),
    ]);
    const report: CaptureReport = {
      schema_version: 1,
      report_integrity_sha256: "0".repeat(64),
      authenticity: authenticity(config),
      source: {
        commit: git.commit,
        dirty: git.dirty,
        files: await hashFiles(root, sourceFiles),
        fixtures: await hashFiles(root, fixtureFiles),
        hooks: await hashFiles(root, hookFiles),
      },
      environment: {
        browser: "chromium",
        browser_version: browser.version(),
        viewport: config.browser.viewport,
        locale: config.browser.locale,
        timezone: config.browser.timezone,
        device_scale_factor: config.browser.device_scale_factor,
        fonts_ready: true,
      },
      adapters: adapters(config),
      map: {
        provider: config.map.provider,
        source_relation: config.map.source_relation,
        execution_mode: config.map.execution_mode,
        ...(config.map.api_key_env === undefined ? {} : { api_key_env: config.map.api_key_env }),
        billing_owner_declared: Boolean(config.map.billing_owner),
        restriction_reviewed: config.map.restriction_reviewed === true,
        readiness: `${config.map.readiness.type}:${config.map.readiness.name}`,
        attribution_visible: attributionVisible,
        map_load_count: mapLoads,
        request_count: externalRequests,
      },
      egress,
      captures: assets,
      ...(pdfReview === undefined ? {} : { pdf: pdfReview }),
      generated_files: [],
    };
    const reportStaging = stagedPath(stagingRoot, config.output.report);
    await fs.mkdir(path.dirname(reportStaging), { recursive: true });
    report.generated_files = generated;
    report.report_integrity_sha256 = reportIntegritySHA256(report);
    assertRedacted(report);
    await validateCaptureReport(report);
    await fs.writeFile(reportStaging, `${JSON.stringify(report, null, 2)}\n`, "utf8");
    const replacementFiles = [
      ...generated,
      { path: config.output.report, sha256: await sha256File(reportStaging) },
    ];
    await replaceGeneratedFiles(root, stagingRoot, replacementFiles, config.output.report);
    return report;
  } finally {
    await context.close();
    await browser.close();
    await runtime.close();
    await fs.rm(stagingRoot, { recursive: true, force: true });
  }
}

async function reviewPDF(
  root: string,
  stagingRoot: string,
  config: CaptureConfiguration,
): Promise<CaptureReport["pdf"] | undefined> {
  if (config.pdf === undefined) return;
  const source = path.join(root, ...config.pdf.path.split("/"));
  const { stdout } = await execFileAsync(
    "python3",
    [
      "-c",
      "from pypdf import PdfReader; import sys; print(len(PdfReader(sys.argv[1]).pages))",
      source,
    ],
    { cwd: root },
  );
  const pageCount = Number.parseInt(stdout.trim(), 10);
  if (!Number.isInteger(pageCount) || pageCount !== config.pdf.expected_pages) {
    throw new Error(
      `PDF page count ${String(pageCount)} does not match expected ${config.pdf.expected_pages}`,
    );
  }
  const rasterDirectory = stagedPath(stagingRoot, config.pdf.raster_output);
  await fs.mkdir(rasterDirectory, { recursive: true });
  await execFileAsync("pdftoppm", ["-png", "-r", "144", source, path.join(rasterDirectory, "page")], {
    cwd: root,
  });
  const pages = (await fs.readdir(rasterDirectory))
    .filter((name) => /^page-[0-9]+\.png$/.test(name))
    .sort((left, right) => left.localeCompare(right, undefined, { numeric: true }));
  if (pages.length !== pageCount) {
    throw new Error("PDF raster review did not produce one PNG per page");
  }
  const rasterPages = await Promise.all(
    pages.map(async (name) => {
      const reportPath = `${config.pdf!.raster_output.replace(/\/$/, "")}/${name}`;
      return {
        path: reportPath,
        sha256: await sha256File(stagedPath(stagingRoot, reportPath)),
      };
    }),
  );
  return {
    path: config.pdf.path,
    sha256: await sha256File(source),
    page_count: pageCount,
    raster_pages: rasterPages,
  };
}

async function installNetworkBoundary(
  context: BrowserContext,
  config: CaptureConfiguration,
  baseURL: string,
  egress: Array<{ method: string; url: string; host: string }>,
  count: () => void,
): Promise<void> {
  const loopbackHost = new URL(baseURL).hostname;
  await context.route("**/*", async (route) => {
    const request = route.request();
    const parsed = new URL(request.url());
    const loopback = ["localhost", "127.0.0.1", "::1", loopbackHost].includes(parsed.hostname);
    if (!loopback && !config.execution.allow_hosts.includes(parsed.hostname)) {
      await route.abort("blockedbyclient");
      return;
    }
    if (!loopback) count();
    egress.push({
      method: request.method(),
      url: redactURL(request.url()),
      host: parsed.hostname,
    });
    await route.continue();
  });
}

async function installHTTPMocks(
  page: Page,
  root: string,
  config: CaptureConfiguration,
): Promise<void> {
  if (config.backend.mode !== "http-route") return;
  for (const routeConfig of config.backend.routes ?? []) {
    await page.route(`**${routeConfig.url}`, async (route) => {
      if (route.request().method().toUpperCase() !== routeConfig.method.toUpperCase()) {
        await route.fallback();
        return;
      }
      const body =
        routeConfig.fixture === undefined
          ? (routeConfig.body ?? "")
          : await fs.readFile(path.join(root, ...routeConfig.fixture.split("/")), "utf8");
      await route.fulfill({
        status: routeConfig.status,
        body,
        contentType: "application/json",
      });
    });
  }
}

async function runStep(
  page: Page,
  step: StepConfig,
  scenarioId: string,
  stepIndex: number,
  stagingRoot: string,
  config: CaptureConfiguration,
): Promise<CaptureAsset | undefined> {
  switch (step.action) {
    case "goto":
      if (step.value === undefined) throw new Error("goto requires a relative value");
      await page.goto(new URL(step.value, page.url()).toString(), { waitUntil: "domcontentloaded" });
      return;
    case "click":
      await locator(page, requiredLocator(step)).click();
      return;
    case "fill":
      await locator(page, requiredLocator(step)).fill(step.value ?? "");
      return;
    case "select":
      await locator(page, requiredLocator(step)).selectOption(step.value ?? "");
      return;
    case "wait":
      await runWait(page, step);
      return;
    case "dialog":
      await runDialog(page, step);
      return;
    case "capture":
      return takeCapture(page, step, scenarioId, stepIndex, stagingRoot, config);
  }
}

async function runWait(page: Page, step: StepConfig): Promise<void> {
  const wait = step.wait;
  if (wait === undefined) throw new Error("wait configuration is absent");
  if (wait.hold_ms !== undefined) {
    await page.waitForTimeout(wait.hold_ms);
    return;
  }
  if (wait.state === "fonts-ready") {
    await page.evaluate(() => document.fonts.ready);
    return;
  }
  if (wait.state === "event") {
    await page.evaluate(
      (name) => new Promise<void>((resolve) => globalThis.addEventListener(name, () => resolve(), { once: true })),
      wait.value ?? "",
    );
    return;
  }
  if (wait.state === "hook") {
    await page.waitForFunction((name) => Boolean((globalThis as Record<string, unknown>)[name]), wait.value ?? "");
    return;
  }
  const target = locator(page, requiredLocator(step));
  if (wait.state === "visible" || wait.state === "hidden") {
    await target.waitFor({ state: wait.state });
  } else if (wait.state === "enabled") {
    await target.click({ trial: true });
  } else if (wait.state === "text") {
    await target.filter({ hasText: wait.value ?? "" }).waitFor();
  } else if (wait.state === "attribute") {
    const expression = wait.value ?? "";
    const separator = expression.indexOf("=");
    if (separator <= 0) throw new Error("attribute wait value must use name=value");
    const name = expression.slice(0, separator);
    const expected = expression.slice(separator + 1);
    await target.evaluate(
      (element, expectedAttribute) =>
        new Promise<void>((resolve, reject) => {
          const matches = (): boolean =>
            element.getAttribute(expectedAttribute.name) === expectedAttribute.value;
          if (matches()) {
            resolve();
            return;
          }
          const observer = new MutationObserver(() => {
            if (!matches()) return;
            clearTimeout(timeout);
            observer.disconnect();
            resolve();
          });
          const timeout = setTimeout(() => {
            observer.disconnect();
            reject(new Error(`attribute ${expectedAttribute.name} did not reach its expected value`));
          }, 30_000);
          observer.observe(element, {
            attributes: true,
            attributeFilter: [expectedAttribute.name],
          });
        }),
      { name, value: expected },
    );
  }
}

async function takeCapture(
  page: Page,
  step: StepConfig,
  scenarioId: string,
  stepIndex: number,
  stagingRoot: string,
  config: CaptureConfiguration,
): Promise<CaptureAsset> {
  const captureConfig = step.capture;
  const captureId = step.id;
  if (captureConfig === undefined || captureId === undefined) throw new Error("capture is incomplete");
  const relative = `${config.output.directory.replace(/\/$/, "")}/${stableAssetName(captureId)}`;
  const absolute = stagedPath(stagingRoot, relative);
  await fs.mkdir(path.dirname(absolute), { recursive: true });
  let clip: { x: number; y: number; width: number; height: number } | undefined;
  if (captureConfig.mode === "locator") {
    const target = locator(page, captureConfig.locator ?? requiredLocator(step));
    const box = await target.boundingBox();
    if (box === null) throw new Error(`capture locator is not visible: ${captureId}`);
    const padding = captureConfig.padding ?? 0;
    const viewport = page.viewportSize();
    if (viewport === null) throw new Error("viewport is unavailable");
    clip = {
      x: Math.max(0, box.x - padding),
      y: Math.max(0, box.y - padding),
      width: Math.min(viewport.width, box.x + box.width + padding) - Math.max(0, box.x - padding),
      height: Math.min(viewport.height, box.y + box.height + padding) - Math.max(0, box.y - padding),
    };
    await page.screenshot({ path: absolute, clip });
  } else if (captureConfig.mode === "region") {
    if (captureConfig.region === undefined) throw new Error("region capture requires bounds");
    clip = captureConfig.region;
    await page.screenshot({ path: absolute, clip });
  } else {
    await page.screenshot({
      path: absolute,
      fullPage: captureConfig.mode === "full-page" || captureConfig.full_page === true,
    });
  }
  if (config.map.provider !== "none" && clip !== undefined) {
    await assertAttributionInClip(page, config, clip);
  }
  const documentationOverlay = await page
    .locator("[data-soku-dialog-overlay]")
    .isVisible()
    .catch(() => false);
  return {
    capture_id: captureId,
    scenario_id: scenarioId,
    step_index: stepIndex,
    path: relative,
    sha256: await sha256File(absolute),
    mode: captureConfig.mode,
    caption: captureConfig.caption,
    ...(clip === undefined ? {} : { clip }),
    ...(documentationOverlay ? { dialog_adapter: "documentation-overlay" } : {}),
  };
}

function locator(page: Page, config: LocatorConfig): Locator {
  switch (config.by) {
    case "role": {
      const [role, name] = config.value.split(":", 2);
      return page.getByRole(role as Parameters<Page["getByRole"]>[0], name === undefined ? {} : { name });
    }
    case "label":
      return page.getByLabel(config.value);
    case "test-id":
      return page.getByTestId(config.value);
    case "text":
      return page.getByText(config.value, { exact: true });
    case "css":
      return page.locator(config.value);
  }
}

function requiredLocator(step: StepConfig): LocatorConfig {
  if (step.locator === undefined) throw new Error(`${step.action} requires a locator`);
  return step.locator;
}

async function waitForFonts(page: Page, fonts: string[]): Promise<void> {
  const ready = await page.evaluate(async (families) => {
    await document.fonts.ready;
    const generic = new Set(["serif", "sans-serif", "monospace", "system-ui"]);
    const sample = "mmmmmmmmmmWWWWWWWWWW日本語😀";
    const canvas = document.createElement("canvas");
    const context = canvas.getContext("2d");
    if (context === null) return false;
    const width = (font: string): number => {
      context.font = `32px ${font}`;
      return context.measureText(sample).width;
    };
    return families.every((family) => {
      if (!document.fonts.check(`16px "${family}"`, "日本語😀")) return false;
      if (generic.has(family.toLowerCase())) return true;
      return ["monospace", "serif", "sans-serif"].some(
        (fallback) => width(`"${family}", ${fallback}`) !== width(fallback),
      );
    });
  }, fonts);
  if (!ready) throw new Error("configured Japanese/emoji font readiness check failed");
}

async function gitIdentity(root: string): Promise<{ commit: string; dirty: boolean }> {
  const [{ stdout: commit }, { stdout: status }] = await Promise.all([
    execFileAsync("git", ["rev-parse", "HEAD"], { cwd: root }),
    execFileAsync("git", ["status", "--porcelain=v1", "--untracked-files=normal"], { cwd: root }),
  ]);
  return { commit: commit.trim(), dirty: status.trim().length > 0 };
}

async function hashFiles(root: string, files: string[]): Promise<HashRecord[]> {
  return Promise.all(
    files.sort().map(async (file) => ({
      path: file,
      sha256: await sha256File(path.join(root, ...file.split("/"))),
    })),
  );
}

function stagedPath(stagingRoot: string, relative: string): string {
  return path.join(stagingRoot, ...relative.split("/"));
}

function unique(values: string[]): string[] {
  return [...new Set(values)].sort();
}

function authenticity(
  config: CaptureConfiguration,
): CaptureReport["authenticity"] {
  return (config.backend.mode === "real-local" || config.backend.mode === "none") &&
    config.map.source_relation !== "declared-adapter" &&
    config.runtime.adapter !== "gas-html-service"
    ? "runtime-authentic"
    : "runtime-authentic-with-adapters";
}

function adapters(config: CaptureConfiguration): string[] {
  const result: string[] = [];
  if (config.backend.mode !== "real-local" && config.backend.mode !== "none") {
    result.push(`backend:${config.backend.mode}`);
  }
  if (config.runtime.adapter === "gas-html-service") result.push("runtime:gas-html-service");
  if (config.map.source_relation === "declared-adapter") result.push(`map:${config.map.provider}`);
  if (config.scenarios.some((scenario) => scenario.steps.some((step) => step.dialog?.mode === "documentation-overlay"))) {
    result.push("dialog:documentation-overlay");
  }
  return result.sort();
}

function renderIndex(config: CaptureConfiguration, assets: CaptureAsset[]): string {
  const title = config.metadata?.manual_title ?? "Generated captures";
  return [
    `# ${title}`,
    "",
    "<!-- Generated by @soku/manual-capture-runner. Edit USAGE.md, not this file. -->",
    "",
    ...assets.flatMap((asset) => [
      `## ${asset.capture_id}`,
      "",
      `![${asset.caption}](./${path.posix.relative(path.posix.dirname(config.output.generated_index), asset.path)})`,
      "",
      `${asset.caption} (${asset.scenario_id}, step ${asset.step_index})`,
      "",
    ]),
  ].join("\n");
}
