# 🧭 Language Selection

> **Applies to:** Both — see [`docs/guides/APPLICABILITY.md`](./APPLICABILITY.md).

## 🎯 Purpose

`docs/guides/STACK_EXAMPLES.md` and `docs/guides/STACK_CONFIGS.md` explain how to configure a stack once it is chosen, but nothing in this boilerplate explains how to choose one. [`INIT_GUIDE.md`](./INIT_GUIDE.md) even tells an AI agent to "ask the user which stack(s) to bootstrap" with no criteria to reason from. This document is that missing criteria: the goal is not to pick the fastest or most popular language, but the one with the lowest total cost across development, deployment, operation, incident response, security, and future hand-off.

## 🧭 What Actually Constrains the Choice

No single global rule governs language choice. In practice, three different kinds of constraints apply, roughly in this priority order:

1. **Platform-forced** — the runtime environment leaves no real choice: browser UI (JavaScript/TypeScript), new Android apps (Kotlin), new iOS apps (Swift), Flutter apps (Dart), database queries (SQL), OS/driver/embedded work (C/C++/Rust), Windows/Microsoft-centric environments (C#/PowerShell).
2. **Organization policy** — a team restricts the option set for operational efficiency (for example: "new backends are Java or Go only," "web frontends must be TypeScript," "shell scripts over 100 lines must be rewritten").
3. **Project criteria** — when the platform doesn't force a choice, weigh compatibility with existing systems, team experience, expected code lifespan, performance/memory needs, security posture, library ecosystem, hiring and hand-off feasibility, deployment/incident-response difficulty, and available test/static-analysis tooling.

## ✅ Questions to Answer Before Choosing

- **Runtime**: Browser, mobile, server, OS/hardware-adjacent, serverless/container, or locked to a specific ecosystem (Windows, JVM, Apple)?
- **Existing systems**: What language does the current codebase use? Must existing auth/payment/DB libraries be reused? Does adding a second language complicate the build/deploy pipeline? Can the current team maintain it?
- **Lifespan and scale**: Hours (throwaway), months (prototype), 3+ years (service), multiple teams touching it, likely to be heavily refactored?
- **Non-functional targets**: Target latency, requests/sec, memory ceiling, cold-start sensitivity, acceptable recovery time, and whether data loss or a security incident would be critical.
- **Operations and staffing**: Can the team debug production incidents in this language? Apply security patches continuously? Hire for it? Keep it running after the original author leaves? Are monitoring/profiling tools available?

## 🚦 Quick Selection Table

| Goal | Language to consider first | Why |
| --- | --- | --- |
| Web frontend | TypeScript | Browser ecosystem, type safety, framework support |
| Small web script | JavaScript | Minimal setup/build overhead |
| Long-lived enterprise backend | Java | Spring ecosystem, long-term maintainability, hiring pool |
| Microsoft-centric backend | C# | .NET, Azure, Windows ecosystem |
| Startup web backend | TypeScript / Python | Fast iteration, rich web ecosystem |
| AI / machine learning | Python | Library and research ecosystem |
| Data analysis | Python + SQL | Analysis libraries plus DB processing |
| Cloud / network services | Go | Concurrency, simple deployment, fast builds |
| Simple ops automation | Bash / PowerShell | Good fit for chaining commands |
| Complex ops automation | Python / Go | Better error handling, testing, structure |
| Android app | Kotlin | Google's Kotlin-first policy |
| iOS app | Swift | Apple platform default |
| Cross-platform app | Dart / Flutter | One UI codebase |
| High-performance engine | C++ | Existing ecosystem, fine-grained performance control |
| Security-critical low-level code | Rust | Memory safety plus performance |
| Embedded | C / C++ / Rust | Hardware access, constrained resources |
| Game development | C++ / C# | Unreal, Unity ecosystems |
| Database processing | SQL | Filtering, aggregation, and joins close to the data |

## 🗂️ Per-Language Fit

| Language | Fits when | Be careful when | Verdict |
| --- | --- | --- | --- |
| JavaScript | Browser-only features, small pages/widgets, quick prototypes, maintaining an existing JS project | Large multi-year projects with several maintainers, complex API data shapes, systems where type errors are costly | Fine for small, short-lived web code; move to TypeScript once it grows |
| TypeScript | React/Next.js/Vue/Nuxt/Angular, long-lived web apps, multi-developer frontends, Node/NestJS backends, shared types across front and back | CPU-heavy numeric/video work, OS/drivers, extreme memory control, one-off scripts a few lines long | Default choice for a new web service |
| Java | Long-running enterprise backends (finance, insurance, public sector, ERP), complex transactions/domain rules, Spring Boot APIs, multi-team systems | Tiny one-off automation, very memory-constrained environments, functions needing extreme cold-start speed, OS kernels/drivers | Very stable choice when long-term maintenance and enterprise systems matter |
| Kotlin | New Android apps, incremental modernization of an existing Java Android app, JVM backends wanting concise syntax, async-heavy apps | A team that only knows Java with no clear reason to add Kotlin, heavy reliance on Java-only codegen tooling | First choice for new Android work; compare against Java by team experience for backends |
| Python | AI/ML, data analysis, ops automation, API integration/data transformation, research, fast prototyping, back-office tools, simple web APIs | Core logic needing extreme CPU performance, services with strict millisecond-level latency budgets, huge codebases with hundreds of maintainers, systems where runtime type errors are critical | First choice for AI, data, automation, and fast validation — mitigate weaknesses with type hints, static type checking, linters, tests, and native libraries at hot paths |
| Go | Network servers, microservices, cloud infrastructure, Kubernetes tooling, CLIs, high-concurrency services, single-binary deploys | AI/data-library-dependent work, complex enterprise business frameworks, tight coupling to an existing Java/Spring system, zero team experience with Go | Strong candidate for cloud-native services and ops tooling |
| C++ | Game engines, database engines, browsers, audio/video processing, high-performance networking, ultra-low-latency systems, existing large C++ codebases | Ordinary CRUD APIs, simple internal systems, early products where speed to launch matters, teams without C++ security/ops experience | Use it when a real performance requirement or existing ecosystem justifies the cost |
| Rust | C/C++-level performance needs, memory safety is critical, cryptography/auth, network processing, OS components, embedded, security-critical server cores, new low-level systems | No team experience, need to ship simple CRUD fast, required SDKs only exist for Java/Python, no actual need for low-level control | First choice for new low-level code where performance and memory safety both matter |
| C# | .NET enterprise systems, Azure-centric environments, Windows desktop, Microsoft 365 integration, ASP.NET Core APIs, Unity games | Org's primary environment is JVM-based, adding a new runtime just for a small project | Natural alternative to Java inside the Microsoft ecosystem |
| PHP | WordPress, CMS, content sites, admin panels, maintaining existing PHP services, Laravel apps | Rewriting an existing PHP system purely because it's out of fashion, forcing new services onto PHP by default | Fits when the existing ecosystem and fast web delivery are the actual goals |
| Ruby | Ruby on Rails, fast web product launches, startup MVPs, maintaining existing Rails systems, productivity-first CRUD | No team Ruby experience, high-throughput is the core requirement, org's standard backend is a different language | Rails ecosystem fit and team experience are the deciding factors |
| Bash / Shell | Chaining a handful of commands, build/deploy wrappers, file moves and simple env setup, small CI/CD steps, small ops utilities | Script exceeds ~100 lines, deeply nested conditionals, complex JSON handling, API retry logic, needs to hold state, many error-recovery paths, needs unit tests, multiple long-term maintainers | Use only for small command chains; move to Python/Go (Linux/cloud) or PowerShell (Windows) once it starts becoming a program |
| SQL | Data queries, filtering, aggregation, joins, transactions, analytical queries — anything the database can do efficiently | Pushing every business rule into stored procedures, making application testing harder, hiding service boundaries inside SQL | Handle data-adjacent operations in SQL; keep service policy and core business rules in application code |
| Dart / Flutter | Shipping Android and iOS from one UI codebase, fast mobile MVPs, delivering the same design across platforms, small teams covering multiple platforms | Needing the latest platform features immediately, platform-specific UX is critical, complex background/hardware integration, large existing native codebase | Decide first whether code sharing or platform optimization matters more |

## ☁️ Infrastructure Languages and Config Formats

Infrastructure work involves configuration languages alongside general-purpose ones:

| Language / format | Primary use |
| --- | --- |
| Bash | Linux command automation |
| PowerShell | Windows, Azure, Active Directory |
| Python | Cloud APIs, ops automation, data processing |
| Go | Infrastructure tooling and platform development |
| HCL | Terraform |
| YAML | Kubernetes, GitHub Actions, GitLab CI |
| Dockerfile | Container image construction |
| SQL | Database operations and analysis |

YAML and HCL are not general-purpose languages for application logic — once a config file accumulates complex conditionals and loops, it becomes hard to maintain. Move that logic into Python, Go, or application code instead.

## 🏛️ Beyond the Style Guide

[`docs/standards/CODE_STYLE.md`](../standards/CODE_STYLE.md) already covers this boilerplate's baseline (Google Style Guide), per-language formatters/linters, and the Google-alignment table — see it for tooling choices. A few selection-relevant principles it doesn't cover:

- **Readers outnumber writers.** Code is written once but read repeatedly by different people over time — prefer several clear lines over one clever line, explicit behavior over implicit, and consistency over personal taste.
- **Static typing's payoff grows with code age and scale.** Short experiments and automation lean toward Python; long-lived enterprise services lean toward statically typed languages (Java, Go, C#, TypeScript for the web layer) because compile-time error detection, safe renames, and large-scale automated refactors matter more as a codebase ages. This is a tendency based on expected lifespan and change volume, not an absolute rule.
- **New low-level code should default to memory-safe languages** (Java, Kotlin, Go, Python, Rust) where the platform allows it. This does not mean rewriting stable existing C++ — apply static analysis, sanitizers, and fuzzing to what already works, and reserve rewrites for genuinely new risk.
- **Shell has a narrow lane**: small wrappers that call other commands, tools with little data manipulation, and nothing where performance matters. Once it grows or gets complex, move to a structured language.

## 🪜 A Practical Selection Procedure

1. **Check the platform.** Web UI, Android, iOS, server, low-level system, or locked to a specific cloud/enterprise ecosystem? If the platform forces a language, the candidate list narrows immediately.
2. **Check the existing codebase.** Current primary language, shared libraries, auth/data-access patterns, build/deploy pipeline, and the operations team's actual experience. Write down the cost of adding a new language.
3. **Decide the expected lifespan.** Hours → writing speed matters most. Weeks → simple structure and docs. Months → tests and deployment. Years → types, API stability, operations, migration path. A decade+ → staffing, compatibility, large-scale automated change, technical debt.
4. **Turn requirements into numbers.** "Should be fast" or "should scale" are not usable requirements. "p95 under 200ms," "1,000 req/s," "under 256MB," "cold start under 1s," "recover within 10 minutes," "zero data loss tolerance" are. Without numbers, performance cannot justify choosing a harder language.
5. **Start from the safest high-level language that meets the requirements.** General automation → Python. Long-lived web service → Java/TypeScript/Go/C#. JVM system → Java/Kotlin. Security-critical low-level code → Rust. Existing high-performance ecosystem → C++.
6. **Confirm the team can actually operate it.** Can they debug an incident, use a profiler, apply security patches, set up tests and CI, keep it running after the original owner leaves, and hire for it?
7. **Record the decision.** See the next section.

## 📝 Recording the Decision

For a choice worth remembering later, write it down rather than leaving the reasoning in someone's memory. This boilerplate has no dedicated ADR (Architecture Decision Record) convention — use the template below inline in an issue, a task report (see [`docs/issues/TASK_REPORT_TEMPLATE.md`](../issues/TASK_REPORT_TEMPLATE.md)), or wherever this repository already records decisions:

```markdown
# Language decision: [project or feature]

- Status: proposed / accepted / superseded
- Date: YYYY-MM-DD

## Context

What problem this project solves and what environment it runs in.

## Requirements

- Functional: ...
- Non-functional: target latency, expected load, memory ceiling,
  availability, security level, expected operating lifespan

## Decision

`Language + main framework + runtime` — e.g. "Java 21, Spring Boot,
Gradle, PostgreSQL, Docker, Cloud Run."

## Why

Team experience, compatibility with existing systems, required
libraries, long-term maintainability, hiring feasibility, ops
tooling, and how it meets the stated requirements.

## Alternatives considered

### Option A
- Pros / cons / why it was not chosen

### Option B
- Pros / cons / why it was not chosen

## Consequences

- Positive: ...
- Negative / cost: ...

## Revisit when

- Load exceeds the original estimate by an order of magnitude
- Team composition changes significantly
- A core library loses support
- Operating cost exceeds target
- Security requirements change
```

## 🧪 Before Adding a New Language

- [ ] There is a concrete problem the current language cannot solve.
- [ ] The new language actually solves that problem.
- [ ] The performance requirement is backed by numbers, not a feeling.
- [ ] The needed libraries/SDKs exist.
- [ ] The build and deploy path is understood.
- [ ] Logging and incident analysis are understood.
- [ ] Security updates can be applied continuously.
- [ ] The team (or you, solo) can maintain it.
- [ ] The communication/data boundary with the existing language is designed.
- [ ] The added cost of introducing this language is written down.
- [ ] At least two alternatives were compared.
- [ ] The decision is recorded (see above).

If most boxes can't be checked, keep the existing language.

## 🚫 Bad Reasons to Pick a Language

Don't choose a language for these reasons alone: it's currently trending, it ranks high on GitHub, a large company uses it, it wins a benchmark, the syntax looks appealing, it would pad out a portfolio's tech list, the existing language "looks old," or another developer recommended it. Tie the choice to the project's actual problem instead.

## 🎬 Summary

Check whether the platform forces the language first. Weight existing systems and team experience most heavily. The longer the expected lifespan and the larger the expected scale, the more static typing and tooling matter. Judge performance by measurement, not intuition. Prefer the safer, more productive language when requirements allow it. Keep Shell to small command chains. Default new low-level code to memory-safe languages, and don't rewrite stable existing code without a reason. Tests, deployment, logging, security, and operations matter more than the language itself. Record every significant choice.
