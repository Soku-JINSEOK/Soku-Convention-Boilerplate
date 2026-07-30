import assert from "node:assert/strict";
import test from "node:test";
import { validateProviderConfiguration } from "./providers.js";
import type { CaptureConfiguration } from "./types.js";

function providerConfiguration(
  provider: CaptureConfiguration["map"]["provider"],
  allowHosts: string[],
): Pick<CaptureConfiguration, "execution" | "map"> {
  return {
    execution: {
      mode: "local-manual",
      allow_hosts: allowHosts,
      request_budget: 16,
    },
    map: {
      provider,
      source_relation: "original-application",
      execution_mode: "local-manual",
      map_load_budget: provider === "google-maps-javascript" ? 1 : 0,
      request_budget: 16,
      readiness: { type: "hook", name: "mapReady" },
      attribution: { required: provider !== "none", locator: ".attribution" },
      ...(provider === "google-maps-javascript"
        ? {
            api_key_env: "GOOGLE_MAPS_API_KEY",
            billing_owner: "declared-owner",
            restriction_reviewed: true,
          }
        : {}),
    },
  };
}

test("provider profiles require exact reviewed hosts", () => {
  assert.doesNotThrow(() =>
    validateProviderConfiguration(providerConfiguration("none", [])),
  );
  assert.doesNotThrow(() =>
    validateProviderConfiguration(
      providerConfiguration("leaflet-osm", ["tile.openstreetmap.org"]),
    ),
  );
  assert.doesNotThrow(() =>
    validateProviderConfiguration(
      providerConfiguration("google-maps-javascript", [
        "maps.googleapis.com",
        "maps.gstatic.com",
      ]),
    ),
  );
  assert.throws(
    () =>
      validateProviderConfiguration(
        providerConfiguration("leaflet-osm", ["example.com"]),
      ),
    /reviewed provider host/,
  );
  assert.throws(
    () =>
      validateProviderConfiguration(
        providerConfiguration("google-maps-javascript", [
          "maps.googleapis.com",
          "maps.gstatic.com",
          "example.com",
        ]),
      ),
    /declarations are incomplete/,
  );
});
