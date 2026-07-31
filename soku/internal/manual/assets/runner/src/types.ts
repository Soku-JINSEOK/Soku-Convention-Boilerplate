export type LocatorConfig = {
  by: "role" | "label" | "test-id" | "text" | "css";
  value: string;
};

export type CaptureConfig = {
  mode: "viewport" | "full-page" | "locator" | "region" | "sequence";
  locator?: LocatorConfig;
  region?: { x: number; y: number; width: number; height: number };
  padding?: number;
  full_page?: boolean;
  caption: string;
};

export type StepConfig = {
  action: "goto" | "click" | "fill" | "select" | "wait" | "capture" | "dialog";
  id?: string;
  locator?: LocatorConfig;
  value?: string;
  wait?: { state: string; value?: string; hold_ms?: number };
  capture?: CaptureConfig;
  dialog?: {
    mode: "native" | "app-owned" | "documentation-overlay";
    action: "accept" | "dismiss";
    message?: string;
  };
};

export type CaptureConfiguration = {
  schema_version: 1;
  execution: {
    mode: "local-manual";
    allow_hosts: string[];
    request_budget: number;
  };
  runtime: {
    adapter: "dev-server" | "static-build" | "gas-html-service";
    command?: string[];
    health_url?: string;
    static_directory?: string;
    source_files?: string[];
    source_fragments?: string[];
    serve: { origin: "loopback-http" };
  };
  browser: {
    engine: "chromium";
    viewport: { width: number; height: number };
    device_scale_factor: number;
    locale: string;
    timezone: string;
    fonts: string[];
  };
  backend: {
    mode: "real-local" | "http-route" | "har-replay" | "browser-bridge" | "none";
    adapter?: string;
    fixture?: string;
    har?: string;
    routes?: Array<{
      method: string;
      url: string;
      status: number;
      fixture?: string;
      body?: string;
    }>;
  };
  map: {
    provider: "none" | "local-deterministic" | "leaflet-osm" | "google-maps-javascript";
    source_relation: "original-application" | "declared-adapter" | "not-applicable";
    api_key_env?: string;
    billing_owner?: string;
    restriction_reviewed?: boolean;
    execution_mode: "local-manual";
    map_load_budget: number;
    request_budget: number;
    readiness: { type: string; name: string };
    attribution: { required: boolean; locator?: string };
  };
  output: {
    directory: string;
    report: string;
    generated_index: string;
  };
  scenarios: Array<{
    id: string;
    route: string;
    steps: StepConfig[];
  }>;
  pdf?: { path: string; expected_pages: number; raster_output: string };
  hooks?: Array<{ name: string; path: string }>;
  metadata?: { manual_title: string };
};

export type HashRecord = { path: string; sha256: string };

export type CaptureAsset = {
  capture_id: string;
  scenario_id: string;
  step_index: number;
  path: string;
  sha256: string;
  mode: string;
  caption: string;
  clip?: { x: number; y: number; width: number; height: number };
  dialog_adapter?: string;
};

export type CaptureReport = {
  schema_version: 1;
  report_integrity_sha256: string;
  authenticity: "runtime-authentic" | "runtime-authentic-with-adapters";
  source: {
    commit: string;
    dirty: boolean;
    files: HashRecord[];
    fixtures: HashRecord[];
    hooks: HashRecord[];
  };
  environment: {
    browser: string;
    browser_version: string;
    viewport: { width: number; height: number };
    locale: string;
    timezone: string;
    device_scale_factor: number;
    fonts_ready: boolean;
  };
  adapters: string[];
  map: {
    provider: string;
    source_relation: string;
    execution_mode: string;
    api_key_env?: string;
    billing_owner_declared: boolean;
    restriction_reviewed: boolean;
    readiness: string;
    attribution_visible: boolean;
    map_load_count: number;
    request_count: number;
  };
  egress: Array<{ method: string; url: string; host: string }>;
  captures: CaptureAsset[];
  pdf?: {
    path: string;
    sha256: string;
    page_count: number;
    raster_pages: HashRecord[];
  };
  generated_files: HashRecord[];
};
