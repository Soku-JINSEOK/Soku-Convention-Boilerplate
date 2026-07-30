# Real-Runtime User-Manual Capture

## Status and Scope

- **Status:** Accepted for the opt-in `docs-manual` component
- **Decision owner:** `Soku-JINSEOK`
- **Decision record:** Issue #164 and its approved task report
- **Execution:** local and manual only

This standard governs screenshots created by running a repository's actual
frontend in Chromium for user-manual evidence. It does not authorize production
access, authentication bypass, customer data, recurring hosted capture, cloud
or billing mutation, operating-system package installation, or manual
publication.

## Authenticity

Every run uses exactly one class:

| Class | Meaning |
| --- | --- |
| `runtime-authentic` | Original frontend and map provider with a real local or test backend and no UI substitution. |
| `runtime-authentic-with-adapters` | Original frontend with disclosed backend, browser-global, fixture, dialog, or map adapters. |
| `illustrative` | A recreated or drawn screen. It is outside this runner and cannot make a real-runtime claim. |

The runner hashes each declared source fragment before launch, keeps adapters
separate, records the Git commit and dirty state, and requires `--allow-dirty`
for a dirty capture. It never rewrites application source.

Provider substitution always changes the class to
`runtime-authentic-with-adapters`. Every affected caption and report must name
the adapter. A documentation overlay for a native browser dialog is also an
adapter and must not be described as captured browser chrome.

## Ownership and Installation

`soku docs manual init --config <path> --dry-run|--yes` installs the versioned
runner package, schemas, example configuration, and manual template as
`core-managed` files. The actual configuration, synthetic fixtures, hooks,
scenario content, PNG files, capture report, generated index, human-authored
prose, and optional PDF remain `project-owned`.

The installer does not create or overwrite `capture.yml`, `USAGE.md`,
translations, fixtures, captures, or PDFs. It rejects existing unowned output,
managed-path collisions, and drift. Apply is transactional and replaces
manifest state last.

The runner uses its own `package-lock.json` and requires Node.js 22 or newer.
Initialization does not run `npm ci`, install a browser, install operating
system packages, or install fonts. `doctor` reports the explicit commands a
reviewer may run.

## Commands

```text
soku docs manual plan --config docs/manual/capture.yml [--json]
soku docs manual doctor --config docs/manual/capture.yml [--probe]
soku docs manual init --config docs/manual/capture.yml --dry-run|--yes
node tools/manual-capture/dist/cli.js capture \
  --config docs/manual/capture.yml [--allow-dirty]
```

Planning is deterministic and read-only. Static doctor checks configuration,
Node and npm, the installed lockfile and built runner, Git state, declared
fonts, output collisions, optional PDF tooling, and environment-variable
presence. It never emits an environment value.

Within the Go CLI, only `doctor --probe` may invoke the installed runner. It
uses fixed argv:

```text
node tools/manual-capture/dist/cli.js probe --config <portable-path>
```

The Go command never invokes a shell or a project command. The installed runner
is invoked directly by the local operator for an explicit capture. It may then
launch the repository-owned `runtime.command` as an argv array with
`shell: false`.

## Configuration and Report Contracts

Capture configuration uses strict schema v1. Unknown fields, duplicate YAML
keys, multiple documents, unsafe paths, literal credentials, unsupported
actions, unsupported providers, unbounded clips, unsanitized HAR files, and
non-local execution are rejected.

The credential-redacted report schema v1 records:

- source, fixture, hook, adapter, image, and generated-file SHA-256 hashes;
- source commit and dirty state;
- Chromium version, viewport, scale, locale, timezone, and font readiness;
- scenario, step, stable capture ID, output path, caption, and final clip;
- backend and dialog adapters;
- redacted external requests and the enforced host allowlist;
- map provider relation, readiness evidence, attribution result, execution
  mode, map-load count, and request count;
- only the Google Maps key environment-variable name, never its value.

Configuration and report schemas are versioned independently from the CLI,
manifest, component catalog, and provider egress profiles.

## Runtime Adapters

### General web applications

`dev-server` launches a reviewed repository command without a shell and waits
for a loopback health URL. `static-build` serves an existing directory on an
ephemeral loopback port. Building the application is an explicit project
action.

`http-route` fulfills declared relative routes from inline or synthetic
fixtures. `har-replay` accepts only a reviewed `*.sanitized.har`; recording a
HAR is never automatic.

### Google Apps Script HTML Service

`gas-html-service` assembles declared source fragments in their configured
order into an ephemeral loopback harness. It does not replace or edit the
fragments. The `google.script.run` bridge preserves chainable success, failure,
and user-object handlers, asynchronous completion, explicit method dispatch,
deterministic error paths, and configured in-memory mutation.

### Dialogs

- `native` records and accepts or dismisses the real dialog without claiming a
  page screenshot contains browser chrome.
- `app-owned` captures the application's real DOM modal.
- `documentation-overlay` injects a disclosed page adapter and records that
  disclosure in the caption and report.

## Capture Behavior

User-facing locators are preferred in this order: role, label, test ID, and
text. CSS is an explicit fallback. Default waits observe visibility, enabled
state, text or attribute state, named events or hooks, responses, map
readiness, and `document.fonts.ready`.

`hold_ms` is bounded to five seconds and reserved for a deliberately transient
manual frame. It is never map readiness.

Capture modes are viewport, full page, locator, region, and sequence. Padded
locator clips are clamped to the viewport. A map capture is refused if the
final clip would crop required attribution.

## Output Replacement

All images, the generated index, and the report are completed and hashed in a
temporary directory. On replacement, the runner reads the previous valid
report, verifies its canonical `report_integrity_sha256`, and verifies every
previously generated asset before touching output. The integrity hash is
computed with that field set to 64 zeroes. The runner refuses an unowned
same-path file or a modified or missing report or generated asset. Replacement
is journaled in a temporary sibling directory, supports rollback, and places
the report last.

Human-authored manual files are never part of generated replacement.

## Network and Provider Policy

Loopback is allowed. All other network access is denied unless the selected
versioned provider profile and configuration both allow the exact hostname.
Unexpected hosts fail rather than widening egress.

### `none`

No map is present. Map loads and provider requests are zero.

### `local-deterministic`

A project-owned deterministic layer is suitable for offline and hosted
synthetic validation. It performs no provider network requests.

### `leaflet-osm`

Live OpenStreetMap public raster tiles are limited to one local/manual viewport,
one declared map load, a bounded request count, visible attribution, normal
browser caching, and no pan/zoom prefetch. The runner does not commit tile bytes
or create an offline cache. Recurring CI must use a provider that permits it or
the deterministic layer.

### `google-maps-javascript`

Live Google Maps JavaScript capture requires all of:

- the original integration or a disclosed adapter;
- a loopback HTTP origin;
- `GOOGLE_MAPS_API_KEY` as the environment-variable name;
- a separately reviewed Website/API-restricted key;
- a declared billing owner, local/manual execution, one map load, and a bounded
  request ceiling;
- a named `idle`, `tilesloaded`, event, or project hook;
- visible Google logo, copyright, and data-provider attribution;
- the versioned egress/CSP profile;
- rejection of key, billing, referer, or provider console failures.

No durable output may contain the key value, a signature, a credential-bearing
URL, cloud project identity, billing status, or personal billing evidence.
`soku` does not create or mutate cloud projects, billing accounts, enabled
APIs, quotas, budgets, keys, restrictions, URL-signing secrets, or App Check.

Maps Static, Street View extraction, tile stitching, satellite repurposing,
prefetching, and offline basemap redistribution are outside this version.

## Fonts and PDF Review

The runner waits for font readiness and checks configured families against
representative Japanese and emoji glyphs. Failure stops capture. Doctor may
recommend fonts but cannot install them.

PDF generation and publication are outside this component. When a project
selects an existing PDF, doctor may require `pdftoppm`; an explicit adapter may
verify the declared page count and rasterize pages for human review.
Rasterization is not a claim that visual layout is correct.

## Hosted and Release Boundary

Hosted validation uses only synthetic fixtures and the deterministic provider.
No live provider key or billable map capture enters CI. PR, scheduled, and
published capture require a separate decision coordinated with Issue #163.

Public CLI, npm, boilerplate, manual, or PDF release remains a separate
release-readiness action under `RELEASE_AND_SYNC.md`.
