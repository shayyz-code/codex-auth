"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const {
  executableNameForPlatform,
  supportedPackages
} = require("./platform");
const { stagePlatformBinaries } = require("./stage-platform-binaries");

test("stages release binaries into platform package bin directories", () => {
  const workspace = fs.mkdtempSync(path.join(os.tmpdir(), "codex-su-stage-"));
  const distDir = path.join(workspace, "dist");
  const packagesDir = path.join(workspace, "packages");
  fs.mkdirSync(distDir, { recursive: true });

  for (const platformPackage of supportedPackages()) {
    fs.writeFileSync(
      path.join(distDir, platformPackage.artifactName),
      platformPackage.packageName
    );
  }

  const staged = stagePlatformBinaries({ distDir, packagesDir });

  assert.equal(staged.length, supportedPackages().length);
  for (const platformPackage of supportedPackages()) {
    const destination = path.join(
      packagesDir,
      platformPackage.packageDir,
      "bin",
      executableNameForPlatform(platformPackage.platform)
    );

    assert.equal(fs.readFileSync(destination, "utf8"), platformPackage.packageName);
    assert.equal(fs.statSync(destination).mode & 0o777, 0o755);
  }
});

test("fails when a release binary is missing", () => {
  const workspace = fs.mkdtempSync(path.join(os.tmpdir(), "codex-su-stage-"));
  const distDir = path.join(workspace, "dist");
  const packagesDir = path.join(workspace, "packages");
  fs.mkdirSync(distDir, { recursive: true });

  assert.throws(
    () => stagePlatformBinaries({ distDir, packagesDir }),
    /Missing release binary:/
  );
});
