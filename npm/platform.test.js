"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");

const {
  executableNameForPlatform,
  packageNameForPlatform,
  supportedPackageNames,
  supportedPackages
} = require("./platform");

test("maps supported npm platforms to release binary packages", () => {
  assert.equal(packageNameForPlatform("darwin", "arm64"), "@shayyz-code/codex-auth-darwin-arm64");
  assert.equal(packageNameForPlatform("darwin", "x64"), "@shayyz-code/codex-auth-darwin-amd64");
  assert.equal(packageNameForPlatform("linux", "arm64"), "@shayyz-code/codex-auth-linux-arm64");
  assert.equal(packageNameForPlatform("linux", "x64"), "@shayyz-code/codex-auth-linux-amd64");
  assert.equal(packageNameForPlatform("win32", "x64"), "@shayyz-code/codex-auth-windows-amd64");
});

test("rejects unsupported npm platforms with a useful error", () => {
  assert.throws(
    () => packageNameForPlatform("freebsd", "x64"),
    /codex-su does not provide an npm binary package for freebsd-x64/
  );
});

test("uses the Windows executable suffix only on Windows", () => {
  assert.equal(executableNameForPlatform("win32"), "codex-su.exe");
  assert.equal(executableNameForPlatform("linux"), "codex-su");
  assert.equal(executableNameForPlatform("darwin"), "codex-su");
});

test("declares every platform package as an optional dependency", () => {
  const packageJSON = require("../package.json");

  assert.deepEqual(
    Object.keys(packageJSON.optionalDependencies).sort(),
    supportedPackageNames()
  );
});

test("defines package metadata for every supported platform package", () => {
  const rootPackageJSON = require("../package.json");

  for (const platformPackage of supportedPackages()) {
    const packageJSON = readPackageJSON(platformPackage.packageDir);

    assert.equal(packageJSON.name, platformPackage.packageName);
    assert.equal(packageJSON.version, rootPackageJSON.version);
    assert.equal(
      rootPackageJSON.optionalDependencies[platformPackage.packageName],
      rootPackageJSON.version
    );
    assert.deepEqual(packageJSON.os, [platformPackage.platform]);
    assert.deepEqual(packageJSON.cpu, [platformPackage.arch]);
    assert.equal(packageJSON.license, rootPackageJSON.license);
    assert.equal(packageJSON.repository.url, rootPackageJSON.repository.url);
    assert.ok(packageJSON.files.includes(`bin/${executableNameForPlatform(platformPackage.platform)}`));
    assert.ok(packageJSON.files.includes("README.md"));
  }
});

function readPackageJSON(packageDir) {
  const packageJSONPath = path.join(__dirname, "packages", packageDir, "package.json");
  return JSON.parse(fs.readFileSync(packageJSONPath, "utf8"));
}
