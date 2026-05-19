"use strict";

const fs = require("node:fs");
const path = require("node:path");

const { supportedPackages } = require("./platform");

const semverPattern = /^(?:v)?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;

function normalizeVersion(rawVersion) {
  const version = String(rawVersion || "").trim();
  const match = version.match(semverPattern);
  if (!match) {
    throw new Error("Version must be a semantic version like 0.2.0 or v0.2.0.");
  }
  return version.replace(/^v/, "");
}

function updateVersion(rawVersion, options = {}) {
  const repoRoot = options.repoRoot || path.join(__dirname, "..");
  const version = normalizeVersion(rawVersion);
  const changed = [];

  const rootPackagePath = path.join(repoRoot, "package.json");
  const rootPackage = readJSON(rootPackagePath);
  rootPackage.version = version;
  rootPackage.optionalDependencies = rootPackage.optionalDependencies || {};

  for (const platformPackage of supportedPackages()) {
    rootPackage.optionalDependencies[platformPackage.packageName] = version;

    const packagePath = path.join(repoRoot, "npm", "packages", platformPackage.packageDir, "package.json");
    const packageJSON = readJSON(packagePath);
    packageJSON.version = version;
    writeJSON(packagePath, packageJSON);
    changed.push(packagePath);
  }

  writeJSON(rootPackagePath, rootPackage);
  changed.unshift(rootPackagePath);

  return { version, changed };
}

function readJSON(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

function writeJSON(filePath, value) {
  fs.writeFileSync(filePath, `${JSON.stringify(value, null, 2)}\n`);
}

if (require.main === module) {
  try {
    const result = updateVersion(process.argv[2]);
    console.log(`Updated npm package versions to ${result.version}.`);
    for (const filePath of result.changed) {
      console.log(`- ${path.relative(path.join(__dirname, ".."), filePath)}`);
    }
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}

module.exports = {
  normalizeVersion,
  updateVersion
};
