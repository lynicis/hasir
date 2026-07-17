# CodeQL Code Scanning Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a CodeQL workflow configuration for GitHub code scanning to analyze Go and TypeScript codebase, ensuring protobuf generation is set up so compilation succeeds.

**Architecture:** Add `.github/workflows/codeql.yml` workflow file. Configure analysis for `go` and `javascript-typescript` languages. Set up Bun, Go, and Buf beforehand to ensure protobuf-generated files are present for the Go compiler/analyzer, preventing build failures.

**Tech Stack:** GitHub Actions, CodeQL, Go, Node/Bun, Buf.

---

### Task 1: Create CodeQL Workflow file

**Files:**
- Create: `.github/workflows/codeql.yml`

**Step 1: Write the CodeQL workflow**
Create the `.github/workflows/codeql.yml` file with standard triggers, permission settings (write permissions for `security-events`), matrix strategy for Go and JavaScript/TypeScript, and pre-build setup steps (like Buf generation) before the CodeQL initialization and autobuild/build actions.

**Step 2: Verify**
Ensure the file has valid YAML syntax.

---
