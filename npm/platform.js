"use strict";

const platformPackages = [
  {
    platform: "darwin",
    arch: "arm64",
    packageName: "@shayyz-code/codex-auth-darwin-arm64",
    packageDir: "codex-auth-darwin-arm64",
    artifactName: "codex-auth-darwin-arm64"
  },
  {
    platform: "darwin",
    arch: "x64",
    packageName: "@shayyz-code/codex-auth-darwin-amd64",
    packageDir: "codex-auth-darwin-amd64",
    artifactName: "codex-auth-darwin-amd64"
  },
  {
    platform: "linux",
    arch: "arm64",
    packageName: "@shayyz-code/codex-auth-linux-arm64",
    packageDir: "codex-auth-linux-arm64",
    artifactName: "codex-auth-linux-arm64"
  },
  {
    platform: "linux",
    arch: "x64",
    packageName: "@shayyz-code/codex-auth-linux-amd64",
    packageDir: "codex-auth-linux-amd64",
    artifactName: "codex-auth-linux-amd64"
  },
  {
    platform: "win32",
    arch: "x64",
    packageName: "@shayyz-code/codex-auth-windows-amd64",
    packageDir: "codex-auth-windows-amd64",
    artifactName: "codex-auth-windows-amd64.exe"
  }
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
    throw new Error(`codex-auth does not provide an npm binary package for ${platform}-${arch}.`);
  }
  return platformPackage.packageName;
}

function executableNameForPlatform(platform = process.platform) {
  return platform === "win32" ? "codex-auth.exe" : "codex-auth";
}

module.exports = {
  executableNameForPlatform,
  packageNameForPlatform,
  supportedPackageNames,
  supportedPackages
};
