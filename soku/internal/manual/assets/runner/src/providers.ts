import type { Page } from "playwright";
import type { CaptureConfiguration } from "./types.js";

type ProviderConfiguration = Pick<CaptureConfiguration, "execution" | "map">;

export function validateProviderConfiguration(
  config: ProviderConfiguration,
): void {
  if (
    config.map.provider === "google-maps-javascript" &&
    (config.map.api_key_env !== "GOOGLE_MAPS_API_KEY" ||
      !config.map.billing_owner ||
      config.map.restriction_reviewed !== true ||
      config.map.map_load_budget !== 1 ||
      config.execution.allow_hosts.join("\0") !==
        ["maps.googleapis.com", "maps.gstatic.com"].join("\0"))
  ) {
    throw new Error("Google Maps live capture declarations are incomplete");
  }
  if (
    config.map.provider === "leaflet-osm" &&
    config.execution.allow_hosts.join("\0") !== "tile.openstreetmap.org"
  ) {
    throw new Error("Leaflet/OSM requires its reviewed provider host");
  }
  if (
    (config.map.provider === "none" ||
      config.map.provider === "local-deterministic") &&
    config.execution.allow_hosts.length !== 0
  ) {
    throw new Error("offline providers require an empty egress allowlist");
  }
}

export async function waitForMap(
  page: Page,
  config: CaptureConfiguration,
): Promise<void> {
  const readiness = config.map.readiness;
  if (readiness.type === "hook") {
    await page.waitForFunction(
      (name) => Boolean((globalThis as Record<string, unknown>)[name]),
      readiness.name,
    );
  } else if (readiness.type === "event") {
    await page.evaluate(
      (name) =>
        new Promise<void>((resolve) =>
          globalThis.addEventListener(name, () => resolve(), { once: true }),
        ),
      readiness.name,
    );
  } else {
    await page.locator(readiness.name).waitFor({ state: "visible" });
  }
}

export async function verifyAttribution(
  page: Page,
  config: CaptureConfiguration,
): Promise<boolean> {
  if (!config.map.attribution.required) return true;
  const selector = config.map.attribution.locator;
  if (selector === undefined) return false;
  const attribution = page.locator(selector);
  await attribution.waitFor({ state: "visible" });
  return attribution.isVisible();
}

export async function assertAttributionInClip(
  page: Page,
  config: CaptureConfiguration,
  clip: { x: number; y: number; width: number; height: number },
): Promise<void> {
  const selector = config.map.attribution.locator;
  if (selector === undefined) {
    throw new Error("map attribution locator is absent");
  }
  const box = await page.locator(selector).boundingBox();
  if (
    box === null ||
    box.x < clip.x ||
    box.y < clip.y ||
    box.x + box.width > clip.x + clip.width ||
    box.y + box.height > clip.y + clip.height
  ) {
    throw new Error("capture clip would crop or obscure required map attribution");
  }
}
