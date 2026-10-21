# soku v0.3.1 release candidate

Release axis: soku

Proposed signed tag: soku/v0.3.1 (not created)

Publication status: unpublished recovery candidate; publication requires exact
source verification and a new human signing/release decision

Manifest schemas: manifest-v1 and manifest-v2 remain byte-compatible and
readable; manifest-v3 supports explicit project-owned core-rendering overrides

Boilerplate compatibility: v1.1.1 candidate; the v1.0.1 through v1.0.5
ownership-handoff coverage and existing lifecycle contracts remain unchanged

Profiles: bootstrap, standard (default and legacy), and scaled

Provider compatibility: provider-v1 with optional deprecated legacy ref;
provider delivery remains separately controlled

Recovery and exit-code contract: unchanged, including manifest-only ownership
handoff and exact previous-manifest rollback

Package matrix: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64

Lifecycle conformance evidence: existing three-OS lifecycle, manifest-v1/v2
regression, manifest-v3 handoff refusal/rollback and package gates remain required

Package and distribution: proposed native and @soku-jinseok/soku 0.3.1;
GitHub Release, archives, checksums and npm delivery remain unverified

Any breaking behavior, read-only compatibility state, or manual recovery
requirement: older CLIs cannot mutate manifest v3; v1/v2 remain compatible

Companion tag: v1.1.1

The immutable public tags `soku/v0.3.0` and `v1.1.0` point to
`772ed6f7aa165da1bbdef89a3129b7703dc5a029`. Their tag-triggered validation
failed before GitHub Release and npm publication: a new tag's all-zero
`before` value was incorrectly used as a Git commit. Both tags and their
historical source records are preserved. This recovery uses new patch versions.

The correction rejects that absent-commit sentinel in the Quick and Security
base inputs. Pull requests retain the trusted base; ordinary branch pushes
retain their actual previous commit. New tags use the tagged commit for both
inputs. Full repository, runtime-template and security gates remain mandatory;
an empty Quick diff does not replace them. No CLI runtime or ownership behavior
is changed by this patch.

Local regression tests evaluate the actual workflow expressions against tag,
branch, pull-request and manual event shapes. An annotated-tag fixture runs the
real scope detector and Quick planner and checks the trusted checkout revision.
These fixtures do not establish actual GitHub tag delivery. Fresh exact-source
hosted validation and successful tag-triggered publication remain required.
