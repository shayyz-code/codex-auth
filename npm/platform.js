"use strict";

const packageByPlatform = {
  "darwin-arm64": "codex-su-darwin-arm64",
  "darwin-x64": "codex-su-darwin-amd64",
  "linux-arm64": "codex-su-linux-arm64",
  "linux-x64": "codex-su-linux-amd64",
  "win32-x64": "codex-su-windows-amd64"
};

function supportedPackageNames() {
  return Object.values(packageByPlatform).sort();
}

function packageNameForPlatform(platform = process.platform, arch = process.arch) {
  const packageName = packageByPlatform[`${platform}-${arch}`];
  if (!packageName) {
    throw new Error(`codex-su does not provide an npm binary package for ${platform}-${arch}.`);
  }
  return packageName;
}

function executableNameForPlatform(platform = process.platform) {
  return platform === "win32" ? "codex-su.exe" : "codex-su";
}

module.exports = {
  executableNameForPlatform,
  packageNameForPlatform,
  supportedPackageNames
};
