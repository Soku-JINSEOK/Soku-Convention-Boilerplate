# 📖 README Guide

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
