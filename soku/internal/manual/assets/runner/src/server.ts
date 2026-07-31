import fs from "node:fs/promises";
import http from "node:http";
import path from "node:path";
import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import type { CaptureConfiguration } from "./types.js";
import { gasBridgeSource, type GasFixture } from "./gas-bridge.js";

export type RunningRuntime = {
  baseURL: string;
  close: () => Promise<void>;
  sourceFiles: string[];
};

export async function startRuntime(
  root: string,
  config: CaptureConfiguration,
): Promise<RunningRuntime> {
  if (config.runtime.adapter === "dev-server") {
    return startDevServer(root, config);
  }
  if (config.runtime.adapter === "gas-html-service") {
    return startGasHarness(root, config);
  }
  return startStaticServer(root, config.runtime.static_directory ?? "");
}

async function startDevServer(
  root: string,
  config: CaptureConfiguration,
): Promise<RunningRuntime> {
  const command = config.runtime.command;
  if (command === undefined || command.length === 0 || config.runtime.health_url === undefined) {
    throw new Error("dev-server configuration is incomplete");
  }
  const process = spawn(command[0]!, command.slice(1), {
    cwd: root,
    shell: false,
    stdio: ["ignore", "pipe", "pipe"],
    env: { ...runtimeEnvironment(config), NODE_ENV: "development" },
  });
  const output: string[] = [];
  process.stdout?.on("data", (chunk: Buffer) => output.push(chunk.toString()));
  process.stderr?.on("data", (chunk: Buffer) => output.push(chunk.toString()));
  await waitForURL(config.runtime.health_url, process, output);
  const parsed = new URL(config.runtime.health_url);
  return {
    baseURL: `${parsed.protocol}//${parsed.host}`,
    close: () => terminate(process),
    sourceFiles: [],
  };
}

async function startStaticServer(root: string, directory: string): Promise<RunningRuntime> {
  const absolute = path.join(root, ...directory.split("/"));
  const sourceFiles = await listFiles(root, absolute);
  const server = http.createServer(async (request, response) => {
    try {
      const requested = new URL(request.url ?? "/", "http://localhost").pathname;
      const relative = requested === "/" ? "index.html" : requested.slice(1);
      const normalized = path.posix.normalize(relative);
      if (normalized.startsWith("../") || path.isAbsolute(normalized)) {
        response.writeHead(400).end();
        return;
      }
      const file = path.join(absolute, ...normalized.split("/"));
      const data = await fs.readFile(file);
      response.setHeader("content-type", contentType(file));
      response.writeHead(200).end(data);
    } catch {
      response.writeHead(404).end("Not found");
    }
  });
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  const address = server.address();
  if (address === null || typeof address === "string") throw new Error("static server failed");
  return {
    baseURL: `http://127.0.0.1:${address.port}`,
    close: async () => {
      server.close();
      await once(server, "close");
    },
    sourceFiles,
  };
}

async function listFiles(root: string, directory: string): Promise<string[]> {
  const result: string[] = [];
  const visit = async (current: string): Promise<void> => {
    const entries = await fs.readdir(current, { withFileTypes: true });
    for (const entry of entries.sort((left, right) => left.name.localeCompare(right.name))) {
      const target = path.join(current, entry.name);
      if (entry.isDirectory()) await visit(target);
      else if (entry.isFile()) result.push(path.relative(root, target).split(path.sep).join("/"));
    }
  };
  await visit(directory);
  return result;
}

async function startGasHarness(
  root: string,
  config: CaptureConfiguration,
): Promise<RunningRuntime> {
  const fragments = config.runtime.source_fragments ?? [];
  const fixturePath = config.backend.fixture;
  if (fixturePath === undefined) throw new Error("GAS browser bridge requires a synthetic fixture");
  const fixture = JSON.parse(
    await fs.readFile(path.join(root, ...fixturePath.split("/")), "utf8"),
  ) as GasFixture;
  const parts = await Promise.all(
    fragments.map(async (fragment) => {
      const content = await fs.readFile(path.join(root, ...fragment.split("/")), "utf8");
      return stripGasTemplateTags(content);
    }),
  );
  const harnessDirectory = await fs.mkdtemp(path.join(root, ".soku-manual-harness-"));
  const overlay = `<script>
    globalThis.__sokuDocumentationDialog = async (message, action) => {
      const overlay = document.createElement("div");
      overlay.setAttribute("data-soku-dialog-overlay", "true");
      overlay.setAttribute("role", "dialog");
      overlay.style.cssText = "position:fixed;inset:0;background:#0008;display:grid;place-items:center;z-index:2147483647";
      overlay.innerHTML = '<div style="background:white;color:#111;padding:24px;border-radius:12px;max-width:32rem">' +
        '<p></p><button type="button">Continue</button></div>';
      overlay.querySelector("p").textContent = message;
      document.body.append(overlay);
      await new Promise((resolve) => overlay.querySelector("button").addEventListener("click", resolve, {once:true}));
      overlay.remove();
      return action === "accept";
    };
  </script>`;
  const html = `<!doctype html><html><head><meta charset="utf-8"></head><body>
${parts.join("\n")}
<script>${gasBridgeSource(fixture)}</script>
${overlay}
</body></html>`;
  await fs.writeFile(path.join(harnessDirectory, "index.html"), html, "utf8");
  const runtime = await startStaticServer(path.dirname(harnessDirectory), path.basename(harnessDirectory));
  return {
    ...runtime,
    sourceFiles: [...fragments],
    close: async () => {
      await runtime.close();
      await fs.rm(harnessDirectory, { recursive: true, force: true });
    },
  };
}

function stripGasTemplateTags(value: string): string {
  return value
    .replace(/<\?!=?[\s\S]*?\?>/g, "")
    .replace(/<html>|<\/html>|<head>|<\/head>|<body>|<\/body>/gi, "");
}

async function waitForURL(
  url: string,
  process: ChildProcess,
  output: string[],
): Promise<void> {
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    if (process.exitCode !== null) {
      throw new Error(`dev server exited before readiness: ${output.join("").slice(-2000)}`);
    }
    try {
      const response = await fetch(url, { signal: AbortSignal.timeout(1000) });
      if (response.ok) return;
    } catch {
      // Readiness is state-based; retry until the bounded deadline.
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  await terminate(process);
  throw new Error("dev server health URL did not become ready");
}

async function terminate(process: ChildProcess): Promise<void> {
  if (process.exitCode !== null) return;
  process.kill("SIGTERM");
  await Promise.race([
    once(process, "exit"),
    new Promise((resolve) => setTimeout(resolve, 5000)),
  ]);
  if (process.exitCode === null) process.kill("SIGKILL");
}

function processEnvWithoutSecrets(): NodeJS.ProcessEnv {
  const allowed = ["PATH", "TMPDIR", "TEMP", "TMP", "SystemRoot", "COMSPEC", "PATHEXT", "HOME"];
  return Object.fromEntries(allowed.flatMap((name) => (process.env[name] === undefined ? [] : [[name, process.env[name]]])));
}

function runtimeEnvironment(config: CaptureConfiguration): NodeJS.ProcessEnv {
  const environment = processEnvWithoutSecrets();
  const keyName = config.map.api_key_env;
  if (
    config.map.provider === "google-maps-javascript" &&
    keyName === "GOOGLE_MAPS_API_KEY" &&
    process.env[keyName] !== undefined
  ) {
    environment[keyName] = process.env[keyName];
  }
  return environment;
}

function contentType(file: string): string {
  if (file.endsWith(".html")) return "text/html; charset=utf-8";
  if (file.endsWith(".js")) return "text/javascript; charset=utf-8";
  if (file.endsWith(".css")) return "text/css; charset=utf-8";
  if (file.endsWith(".json")) return "application/json; charset=utf-8";
  if (file.endsWith(".svg")) return "image/svg+xml";
  if (file.endsWith(".png")) return "image/png";
  return "application/octet-stream";
}
