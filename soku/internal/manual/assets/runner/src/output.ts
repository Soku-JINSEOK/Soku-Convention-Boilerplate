import crypto from "node:crypto";
import fs from "node:fs/promises";
import path from "node:path";
import type { CaptureReport, HashRecord } from "./types.js";

export async function sha256File(filePath: string): Promise<string> {
  const data = await fs.readFile(filePath);
  return crypto.createHash("sha256").update(data).digest("hex");
}

export function stableAssetName(captureId: string): string {
  if (!/^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$/.test(captureId)) {
    throw new Error(`invalid stable capture id: ${captureId}`);
  }
  return `${captureId}.png`;
}

export function reportIntegritySHA256(report: CaptureReport): string {
  const canonical = {
    ...report,
    report_integrity_sha256: "0".repeat(64),
  };
  return crypto
    .createHash("sha256")
    .update(`${JSON.stringify(canonical, null, 2)}\n`)
    .digest("hex");
}

export async function replaceGeneratedFiles(
  root: string,
  stagingRoot: string,
  nextFiles: HashRecord[],
  reportPath: string,
): Promise<void> {
  const reportAbsolute = path.join(root, ...reportPath.split("/"));
  let previous: CaptureReport | undefined;
  try {
    previous = JSON.parse(await fs.readFile(reportAbsolute, "utf8")) as CaptureReport;
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code !== "ENOENT") {
      throw error;
    }
  }
  if (
    previous !== undefined &&
    previous.report_integrity_sha256 !== reportIntegritySHA256(previous)
  ) {
    throw new Error(`generated output was modified or removed: ${reportPath}`);
  }
  const previousFiles = new Map(
    previous?.generated_files.map((item) => [item.path, item.sha256]) ?? [],
  );
  for (const item of previous?.generated_files ?? []) {
    const absolute = path.join(root, ...item.path.split("/"));
    const currentHash = await sha256File(absolute).catch(() => "");
    if (currentHash !== item.sha256) {
      throw new Error(`generated output was modified or removed: ${item.path}`);
    }
  }
  for (const item of nextFiles) {
    const absolute = path.join(root, ...item.path.split("/"));
    if (!previousFiles.has(item.path) && !(item.path === reportPath && previous !== undefined)) {
      try {
        await fs.lstat(absolute);
        throw new Error(`refusing to replace an unowned output: ${item.path}`);
      } catch (error) {
        if ((error as NodeJS.ErrnoException).code !== "ENOENT") {
          throw error;
        }
      }
    }
  }
  await fs.mkdir(path.dirname(reportAbsolute), { recursive: true });
  const transaction = await fs.mkdtemp(
    path.join(path.dirname(reportAbsolute), ".replace-"),
  );
  const replaced: Array<{ path: string; backup?: string }> = [];
  try {
    for (const previousPath of previousFiles.keys()) {
      if (nextFiles.some((item) => item.path === previousPath)) continue;
      const target = path.join(root, ...previousPath.split("/"));
      const backup = path.join(transaction, crypto.randomUUID());
      await fs.rename(target, backup);
      replaced.push({ path: target, backup });
    }
    const ordered = [...nextFiles].sort((left, right) => {
      if (left.path === reportPath) return 1;
      if (right.path === reportPath) return -1;
      return left.path.localeCompare(right.path);
    });
    for (const item of ordered) {
      const target = path.join(root, ...item.path.split("/"));
      const source = path.join(stagingRoot, ...item.path.split("/"));
      await fs.mkdir(path.dirname(target), { recursive: true });
      let backup: string | undefined;
      try {
        await fs.lstat(target);
        backup = path.join(transaction, crypto.randomUUID());
        await fs.rename(target, backup);
      } catch (error) {
        if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
      }
      await fs.rename(source, target);
      replaced.push({ path: target, ...(backup === undefined ? {} : { backup }) });
    }
  } catch (error) {
    for (const item of replaced.reverse()) {
      await fs.rm(item.path, { force: true });
      if (item.backup !== undefined) await fs.rename(item.backup, item.path);
    }
    throw error;
  } finally {
    await fs.rm(transaction, { recursive: true, force: true });
  }
}
