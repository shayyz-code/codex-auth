const assert = require("node:assert/strict");
const fs = require("node:fs");
const fsp = require("node:fs/promises");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const root = path.resolve(__dirname, "..");

function clearDistModule(relativePath) {
  delete require.cache[require.resolve(path.join(root, relativePath))];
}

function loadAccountService(codexHome) {
  process.env.CODEX_HOME = codexHome;
  clearDistModule("dist/lib/config/paths.js");
  clearDistModule("dist/lib/accounts/account-service.js");
  return require(path.join(root, "dist/lib/accounts/account-service.js")).AccountService;
}

async function withCodexHome(fn) {
  const codexHome = await fsp.mkdtemp(path.join(os.tmpdir(), "codex-su-test-"));
  try {
    return await fn(codexHome);
  } finally {
    delete process.env.CODEX_HOME;
    await fsp.rm(codexHome, { recursive: true, force: true });
  }
}

async function writeAuth(codexHome, value) {
  await fsp.mkdir(codexHome, { recursive: true });
  await fsp.writeFile(path.join(codexHome, "auth.json"), JSON.stringify(value), "utf8");
}

test("saveAccount snapshots auth.json and listAccountNames sorts saved accounts", async () => {
  await withCodexHome(async (codexHome) => {
    const AccountService = loadAccountService(codexHome);
    const service = new AccountService();

    await writeAuth(codexHome, { token: "alpha" });
    assert.equal(await service.saveAccount("Beta.json"), "Beta");

    await writeAuth(codexHome, { token: "bravo" });
    assert.equal(await service.saveAccount("alpha"), "alpha");

    assert.deepEqual(await service.listAccountNames(), ["alpha", "Beta"]);
    assert.equal(
      await fsp.readFile(path.join(codexHome, "accounts", "Beta.json"), "utf8"),
      JSON.stringify({ token: "alpha" }),
    );
  });
});

test("useAccount activates a saved account and records the current account", async () => {
  await withCodexHome(async (codexHome) => {
    const AccountService = loadAccountService(codexHome);
    const service = new AccountService();

    await fsp.mkdir(path.join(codexHome, "accounts"), { recursive: true });
    await fsp.writeFile(
      path.join(codexHome, "accounts", "work.json"),
      JSON.stringify({ token: "work" }),
      "utf8",
    );

    assert.equal(await service.useAccount("work"), "work");
    assert.equal(await service.getCurrentAccountName(), "work");
    assert.equal(
      await fsp.readFile(path.join(codexHome, "auth.json"), "utf8"),
      JSON.stringify({ token: "work" }),
    );

    if (process.platform !== "win32") {
      assert.equal(fs.lstatSync(path.join(codexHome, "auth.json")).isSymbolicLink(), true);
    }
  });
});

test("getCurrentAccountName can infer the account from an auth.json symlink", async () => {
  if (process.platform === "win32") {
    return;
  }

  await withCodexHome(async (codexHome) => {
    const AccountService = loadAccountService(codexHome);
    const service = new AccountService();
    const accountPath = path.join(codexHome, "accounts", "personal.json");

    await fsp.mkdir(path.dirname(accountPath), { recursive: true });
    await fsp.writeFile(accountPath, JSON.stringify({ token: "personal" }), "utf8");
    await fsp.symlink(accountPath, path.join(codexHome, "auth.json"));

    assert.equal(await service.getCurrentAccountName(), "personal");
  });
});

test("saveAccount rejects invalid names and missing auth.json", async () => {
  await withCodexHome(async (codexHome) => {
    const AccountService = loadAccountService(codexHome);
    const service = new AccountService();

    await assert.rejects(() => service.saveAccount("../bad"), {
      name: "InvalidAccountNameError",
    });
    await assert.rejects(() => service.saveAccount("missing-auth"), {
      name: "AuthFileMissingError",
    });
  });
});
