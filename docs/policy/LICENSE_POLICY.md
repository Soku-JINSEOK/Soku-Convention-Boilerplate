# 📜 License Policy

## 🎯 Purpose

This document defines how repositories based on `Soku-Convention-Boilerplate` should approach licensing decisions.

Licensing is not just a legal checkbox.  
It affects reuse, contribution flow, company adoption, internal distribution, and long-term project positioning.

## 📐 Core Principles

License decisions should be:

- explicit
- documented
- compatible with intended usage
- reviewed before external distribution

## ✅ Default Expectation

Every repository should declare its license clearly.  
At minimum, this should include:

- a root-level `LICENSE` file
- a short license note in `README.md`
- documentation of any third-party license constraints when relevant

## 🤔 How to Choose a License

Choose a license based on the real operating model of the repository.

Questions to answer:

- Is the project private, internal, public, or commercial?
- Should outside users be able to modify and redistribute it?
- Are there patent or enterprise adoption concerns?
- Will the project be used as a reusable starter across teams or organizations?

## 🛠️ Practical Guidance

### 🟢 MIT

Choose `MIT` when broad reuse, simplicity, and low friction matter most.  
This is often a strong default for open boilerplates and developer templates.

### 🔵 Apache-2.0

Choose `Apache-2.0` when you want permissive reuse plus clearer patent language.  
This is often preferred in company or platform contexts where legal clarity matters.

For this boilerplate repository, `MIT` is the recommended default because it keeps reuse friction low for downstream projects while remaining easy to understand.

### 🔒 Proprietary / Internal

Choose an internal or proprietary license model when the repository contains company-specific assets, internal operational logic, or restricted business value.

## 📦 Third-Party Dependencies

Repositories should track material third-party license obligations when dependencies introduce:

- attribution requirements
- copyleft implications
- redistribution restrictions
- patent-related conditions

## 🔍 Review Rule

If a repository is intended for public release, license choice should be made deliberately before publication rather than added as an afterthought.

## 🎬 Summary

The right license supports the real use case of the repository.  
Good license policy reduces future friction around reuse, contribution, and distribution.
