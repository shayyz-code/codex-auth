#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");

const {
  executableNameForPlatform,
  supportedPackages
} = require("./platform");

function stagePlatformBinaries(options = {}) {
  const repoRoot = options.repoRoot || path.join(__dirname, "..");
  const distDir = options.distDir || path.join(repoRoot, "dist");
  const packagesDir = options.packagesDir || path.join(__dirname, "packages");
  const staged = [];

  for (const platformPackage of supportedPackages()) {
    const source = path.join(distDir, platformPackage.artifactName);
    const destination = path.join(
      packagesDir,
      platformPackage.packageName,
      "bin",
      executableNameForPlatform(platformPackage.platform)
    );

    if (!fs.existsSync(source)) {
      throw new Error(`Missing release binary: ${source}`);
    }

    fs.mkdirSync(path.dirname(destination), { recursive: true });
    fs.copyFileSync(source, destination);
    fs.chmodSync(destination, 0o755);
    staged.push({ source, destination });
  }

  return staged;
}

function main() {
  try {
    const staged = stagePlatformBinaries();
    for (const { source, destination } of staged) {
      console.log(`${source} -> ${destination}`);
    }
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = {
  stagePlatformBinaries
};
