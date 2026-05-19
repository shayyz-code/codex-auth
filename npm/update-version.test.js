"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const {
  supportedPackages
} = require("./platform");
const {
  normalizeVersion,
  updateVersion
} = require("./update-version");

test("normalizes semantic versions with an optional v prefix", () => {
  assert.equal(normalizeVersion("0.2.0"), "0.2.0");
  assert.equal(normalizeVersion("v0.2.0"), "0.2.0");
});

test("rejects invalid versions", () => {
  assert.throws(
    () => normalizeVersion("latest"),
    /Version must be a semantic version/
  );
});

test("updates root and platform npm package versions", () => {
  const workspace = fs.mkdtempSync(path.join(os.tmpdir(), "codex-auth-version-"));
  copyPackageFixture(workspace);

  const result = updateVersion("v0.2.0", { repoRoot: workspace });

  assert.equal(result.version, "0.2.0");

  const rootPackage = readJSON(path.join(workspace, "package.json"));
  assert.equal(rootPackage.version, "0.2.0");

  for (const platformPackage of supportedPackages()) {
    assert.equal(rootPackage.optionalDependencies[platformPackage.packageName], "0.2.0");

    const packageJSON = readJSON(
      path.join(workspace, "npm", "packages", platformPackage.packageDir, "package.json")
    );
    assert.equal(packageJSON.version, "0.2.0");
  }
});

function copyPackageFixture(workspace) {
  fs.copyFileSync(
    path.join(__dirname, "..", "package.json"),
    path.join(workspace, "package.json")
  );

  for (const platformPackage of supportedPackages()) {
    const source = path.join(__dirname, "packages", platformPackage.packageDir, "package.json");
    const destination = path.join(workspace, "npm", "packages", platformPackage.packageDir, "package.json");
    fs.mkdirSync(path.dirname(destination), { recursive: true });
    fs.copyFileSync(source, destination);
  }
}

function readJSON(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}
