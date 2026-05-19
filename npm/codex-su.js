#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const {
  executableNameForPlatform,
  packageNameForPlatform
} = require("./platform");

function main() {
  let packageName;
  try {
    packageName = packageNameForPlatform();
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }

  let packageJSONPath;
  try {
    packageJSONPath = require.resolve(`${packageName}/package.json`);
  } catch {
    console.error(
      `Missing optional dependency ${packageName}. Reinstall codex-su for this platform.`
    );
    process.exit(1);
  }

  const executablePath = path.join(
    path.dirname(packageJSONPath),
    "bin",
    executableNameForPlatform()
  );
  const result = spawnSync(executablePath, process.argv.slice(2), {
    stdio: "inherit"
  });

  if (result.error) {
    console.error(result.error.message);
    process.exit(1);
  }
  process.exit(result.status === null ? 1 : result.status);
}

main();
