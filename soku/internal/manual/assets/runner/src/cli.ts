#!/usr/bin/env node
import path from "node:path";
import process from "node:process";
import { capture, probe } from "./capture.js";
import { loadConfiguration } from "./config.js";
import { redactText } from "./redaction.js";

type Arguments = {
  command: "capture" | "probe";
  config: string;
  allowDirty: boolean;
};

function parseArguments(values: string[]): Arguments {
  const command = values[0];
  if (command !== "capture" && command !== "probe") {
    throw new Error("usage: manual-capture <capture|probe> --config <path> [--allow-dirty]");
  }
  let config = "";
  let allowDirty = false;
  for (let index = 1; index < values.length; index += 1) {
    const value = values[index];
    if (value === "--config") {
      config = values[index + 1] ?? "";
      index += 1;
    } else if (value === "--allow-dirty" && command === "capture") {
      allowDirty = true;
    } else {
      throw new Error(`unknown argument: ${value}`);
    }
  }
  if (config === "") throw new Error("--config is required");
  return { command, config, allowDirty };
}

async function main(): Promise<void> {
  const args = parseArguments(process.argv.slice(2));
  const root = path.resolve(process.cwd());
  if (args.command === "probe") {
    process.stdout.write(`${await probe(root, args.config)}\n`);
    return;
  }
  const configuration = await loadConfiguration(root, args.config);
  const report = await capture(root, args.config, args.allowDirty);
  process.stdout.write(
    `${JSON.stringify({
      ok: true,
      authenticity: report.authenticity,
      captures: report.captures.length,
      report: configuration.output.report,
    })}\n`,
  );
}

main().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`manual-capture: ${redactText(message)}\n`);
  process.exitCode = 1;
});
