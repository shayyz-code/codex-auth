"use strict";

const platformPackages = [
  { platform: "darwin", arch: "arm64", packageName: "codex-su-darwin-arm64" },
  { platform: "darwin", arch: "x64", packageName: "codex-su-darwin-amd64" },
  { platform: "linux", arch: "arm64", packageName: "codex-su-linux-arm64" },
  { platform: "linux", arch: "x64", packageName: "codex-su-linux-amd64" },
  { platform: "win32", arch: "x64", packageName: "codex-su-windows-amd64" }
];

function supportedPackageNames() {
  return supportedPackages().map(({ packageName }) => packageName).sort();
}

function supportedPackages() {
  return platformPackages.map((platformPackage) => ({ ...platformPackage }));
}

function packageNameForPlatform(platform = process.platform, arch = process.arch) {
  const platformPackage = platformPackages.find(
    (candidate) => candidate.platform === platform && candidate.arch === arch
  );
  if (!platformPackage) {
    throw new Error(`codex-su does not provide an npm binary package for ${platform}-${arch}.`);
  }
  return platformPackage.packageName;
}

function executableNameForPlatform(platform = process.platform) {
  return platform === "win32" ? "codex-su.exe" : "codex-su";
}

module.exports = {
  executableNameForPlatform,
  packageNameForPlatform,
  supportedPackageNames,
  supportedPackages
};
