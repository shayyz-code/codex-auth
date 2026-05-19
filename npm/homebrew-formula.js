#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");

const { supportedPackages } = require("./platform");

const owner = "shayyz-code";
const repo = "codex-auth";
const homepage = `https://github.com/${owner}/${repo}`;

function generateHomebrewFormula(options) {
  const tag = options.tag;
  const distDir = options.distDir;
  if (!tag) {
    throw new Error("A release tag is required.");
  }
  if (!distDir) {
    throw new Error("A dist directory is required.");
  }

  const version = tag.replace(/^v/, "");
  const checksums = readChecksums(distDir);

  return `class CodexAuth < Formula
  desc "Manage named Codex auth snapshots"
  homepage "${homepage}"
  version "${version}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "${releaseURL(tag, "codex-auth-darwin-arm64")}"
      sha256 "${checksums.get("codex-auth-darwin-arm64")}"
    else
      url "${releaseURL(tag, "codex-auth-darwin-amd64")}"
      sha256 "${checksums.get("codex-auth-darwin-amd64")}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "${releaseURL(tag, "codex-auth-linux-arm64")}"
      sha256 "${checksums.get("codex-auth-linux-arm64")}"
    else
      url "${releaseURL(tag, "codex-auth-linux-amd64")}"
      sha256 "${checksums.get("codex-auth-linux-amd64")}"
    end
  end

  def install
    bin.install Dir["codex-auth-*"].first => "codex-auth"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/codex-auth --version")
  end
end
`;
}

function readChecksums(distDir) {
  const requiredArtifacts = supportedPackages()
    .filter(({ platform }) => platform !== "win32")
    .map(({ artifactName }) => artifactName);
  const checksums = new Map();

  for (const artifactName of requiredArtifacts) {
    const checksumPath = path.join(distDir, `${artifactName}.sha256`);
    if (!fs.existsSync(checksumPath)) {
      throw new Error(`Missing checksum file: ${checksumPath}`);
    }

    const checksum = fs.readFileSync(checksumPath, "utf8").trim().split(/\s+/)[0];
    if (!/^[a-f0-9]{64}$/i.test(checksum)) {
      throw new Error(`Invalid sha256 in ${checksumPath}`);
    }
    checksums.set(artifactName, checksum.toLowerCase());
  }

  return checksums;
}

function releaseURL(tag, artifactName) {
  return `${homepage}/releases/download/${tag}/${artifactName}`;
}

function main() {
  try {
    const tag = process.env.GITHUB_REF_NAME || process.argv[2];
    const distDir = process.argv[3] || path.join(__dirname, "..", "dist");
    process.stdout.write(generateHomebrewFormula({ tag, distDir }));
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = {
  generateHomebrewFormula,
  readChecksums
};
