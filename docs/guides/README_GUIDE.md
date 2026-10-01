# 📖 README Guide

> **Document purpose:** Document presentation guide. Defines what readers should learn immediately from a README and how to navigate deeper documentation.
>
> **Key point:** Every document needs a clear top summary; diagrams explain relationships and prose explains the decision.

## First-screen contract for every document

Apply this presentation rule to READMEs, standards, policies, guides, templates,
infrastructure instructions and task reports. Keep the existing language policy.

1. Immediately below the title, state the document's role, what question it
   answers and its key conclusion. Keep language-navigation links near the top.
2. For a structural or behavioral topic, add a scoped diagram near its overview.
   Show roles, significant inputs/outputs and decision or failure paths when
   they belong to the subject. Label reference patterns and future designs
   explicitly; do not imply that a diagram proves deployed infrastructure.
3. For prose, policy, examples or a historical report, a specific summary is
   sufficient. Do not force a diagram onto a topic without useful relationships.
4. Explain how to read each diagram and include a concrete reader-check question.
   Define unfamiliar terms in nearby prose. Readers must understand the main
   conclusion even when the diagram cannot render.
5. Use one relationship or scenario per diagram, short labels and a top-down
   layout when needed. Limit horizontal breadth to five nodes; split larger
   views. Distinguish control/data flow from configuration or evidence.
6. Keep historical status and actual test results intact. Versioned release
   records and append-only ledgers retain their original content; provide
   summaries in the adjacent release index.
7. Maintain the [complete document map](./DOCUMENTATION_MAP.md) when adding,
   moving or retiring a document.

### Why this format

Readers need to establish relevance before learning implementation detail.
The purpose block answers whether the document is the right one; the diagram
makes responsibilities and branching visible; prose preserves the reasons,
limits and operating instructions. The accepted maintenance cost is keeping
the summary, diagram and behavior consistent in the same change.

Before review, ask: can a first-time reader identify this document's role,
the responsible component, the normal result and the next action after failure?
For descriptive records, can they identify the topic and the evidence date?

## 🎯 Purpose

This document defines how README files should be managed in repositories based on `Soku-Convention-Boilerplate`.

The README is the front door of the repository — the first thing a stranger (or an AI agent) sees before anything else.  
It should help a contributor understand what the project is, why it exists, how it is used, and where to go next.

## 🚪 Role of the README

A good README should answer the first questions a contributor is likely to have:

- What is this repository for?
- What problem does it solve?
- How do I get started?
- What standards does it follow?
- Where can I find deeper documentation?

## 🎨 Tone and Presentation

README files should feel clear, modern, and intentional.  
They do not need to be flashy, but they should avoid looking like an unstructured dump of notes.

Prefer:

- strong sectioning
- concise lead-in text
- consistent heading hierarchy
- short tables where they improve scanning
- example-driven explanation

## 🌐 Language Policy

See the [Language Policy in BLUEPRINT.md](../../BLUEPRINT.md#language-policy) for the single canonical rule on document language, including the Multi-Language Block Ordering rule: when a README mixes languages, group each language into one contiguous block (English first, then each additional language in turn) instead of interleaving them section by section. See `README.md` for a reference implementation.

## 🗂️ Recommended README Structure

```text
1. Project title
2. Short value statement
3. Overview
4. Why this project exists
5. Key standards or principles
6. Getting started
7. Documentation map
8. Stack or capability summary
9. Contribution entry points
```

## 🚫 What to Avoid

Avoid README files that are:

- too shallow to be useful
- too long without structure
- full of outdated setup instructions
- inconsistent with actual repository behavior

## 🔁 Maintenance Rule

The README should be updated whenever repository behavior, setup flow, or core positioning changes materially.

If the repository changes but the README stays frozen, onboarding quality degrades quickly.

## 🧭 Documentation Map

The README should act as a hub, not as the only document.  
It should point clearly to:

- contribution rules
- code style guidance
- CI/CD documentation
- architecture or design references
- agent instructions

## 🎬 Summary

Treat the README as product-quality documentation for the repository itself.  
A strong README reduces onboarding friction for both humans and AI agents before they ever inspect the code.
