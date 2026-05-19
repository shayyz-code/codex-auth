"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const { generateHomebrewFormula } = require("./homebrew-formula");

test("generates a Homebrew formula for macOS and Linux release binaries", () => {
  const distDir = createDistWithChecksums();

  const formula = generateHomebrewFormula({ tag: "v1.2.3", distDir });

  assert.match(formula, /class CodexAuth < Formula/);
  assert.match(formula, /version "1\.2\.3"/);
  assert.match(formula, /license "MIT"/);
  assert.match(formula, /on_macos do/);
  assert.match(formula, /on_linux do/);
  assert.match(formula, /Hardware::CPU\.arm\?/);
  assert.match(formula, /codex-auth-darwin-arm64/);
  assert.match(formula, /codex-auth-darwin-amd64/);
  assert.match(formula, /codex-auth-linux-arm64/);
  assert.match(formula, /codex-auth-linux-amd64/);
  assert.doesNotMatch(formula, /windows/);
  assert.match(formula, /bin\.install Dir\["codex-auth-\*"\]\.first => "codex-auth"/);
  assert.match(formula, /assert_match version\.to_s/);
});

test("fails when a required Homebrew checksum is missing", () => {
  const distDir = fs.mkdtempSync(path.join(os.tmpdir(), "codex-auth-formula-"));

  assert.throws(
    () => generateHomebrewFormula({ tag: "v1.2.3", distDir }),
    /Missing checksum file:/
  );
});

function createDistWithChecksums() {
  const distDir = fs.mkdtempSync(path.join(os.tmpdir(), "codex-auth-formula-"));
  const artifacts = [
    "codex-auth-darwin-arm64",
    "codex-auth-darwin-amd64",
    "codex-auth-linux-arm64",
    "codex-auth-linux-amd64"
  ];

  for (const [index, artifact] of artifacts.entries()) {
    const checksum = String(index + 1).repeat(64);
    fs.writeFileSync(path.join(distDir, `${artifact}.sha256`), `${checksum}  ${artifact}\n`);
  }

  return distDir;
}
