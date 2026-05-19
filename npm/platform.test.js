"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");

const {
  executableNameForPlatform,
  packageNameForPlatform,
  supportedPackageNames
} = require("./platform");

test("maps supported npm platforms to release binary packages", () => {
  assert.equal(packageNameForPlatform("darwin", "arm64"), "codex-su-darwin-arm64");
  assert.equal(packageNameForPlatform("darwin", "x64"), "codex-su-darwin-amd64");
  assert.equal(packageNameForPlatform("linux", "arm64"), "codex-su-linux-arm64");
  assert.equal(packageNameForPlatform("linux", "x64"), "codex-su-linux-amd64");
  assert.equal(packageNameForPlatform("win32", "x64"), "codex-su-windows-amd64");
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
