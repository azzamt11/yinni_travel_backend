const http = require("http");
const fs = require("fs");
const path = require("path");
const yaml = require("js-yaml");

const { AuthAPIService } = require("./auth-api.service");
const { jwt } = require("./middleware/jwt");
const { extractClaimsFromContext } = require("./middleware/claims");

function getArg(name, fallback) {
  const idx = process.argv.indexOf(name);
  if (idx >= 0 && idx + 1 < process.argv.length) {
    return process.argv[idx + 1];
  }
  return fallback;
}

function loadConfig(confArg) {
  const confPath = confArg || "./configs";
  const filePath = fs.statSync(confPath).isDirectory()
    ? path.join(confPath, "config.yaml")
    : confPath;

  const raw = fs.readFileSync(filePath, "utf8");
  const cfg = yaml.load(raw) || {};
  const jwtExpireRaw = cfg?.auth?.jwt_expired ?? cfg?.auth?.jwt_expire ?? 3600;
  const httpAddrRaw = cfg?.server?.http?.addr || "0.0.0.0:8000";
  const jwtSecretRaw = cfg?.auth?.jwt_secret || "dev-secret";

  return {
    addr: resolveEnvExpr(String(httpAddrRaw)),
    jwtSecret: resolveEnvExpr(String(jwtSecretRaw)),
    jwtExpired: Number(resolveEnvExpr(String(jwtExpireRaw))),
  };
}

function resolveEnvExpr(value) {
  // Supports "${ENV:default}" used in kratos-style config.
  const m = value.match(/^\$\{([^:}]+):([^}]*)\}$/);
  if (!m) return value;
  const envName = m[1];
  const fallback = m[2];
  const envVal = process.env[envName];
  return envVal && envVal.length > 0 ? envVal : fallback;
}

function parseAddr(addr) {
  const i = addr.lastIndexOf(":");
  if (i < 0) return { host: "0.0.0.0", port: 8000 };
  return { host: addr.slice(0, i), port: Number(addr.slice(i + 1)) || 8000 };
}

async function readJSON(req) {
  return new Promise((resolve, reject) => {
    let data = "";
    req.on("data", (chunk) => (data += chunk));
    req.on("end", () => {
      try {
        resolve(data ? JSON.parse(data) : {});
      } catch {
        reject(new Error("invalid json"));
      }
    });
    req.on("error", reject);
  });
}

function writeJSON(res, status, body) {
  const text = JSON.stringify(body || {});
  res.writeHead(status, {
    "Content-Type": "application/json",
    "Content-Length": Buffer.byteLength(text),
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Headers": "Content-Type, Authorization",
    "Access-Control-Allow-Methods": "POST, GET, OPTIONS",
  });
  res.end(text);
}

async function main() {
  const conf = getArg("-conf", "./configs");
  const cfg = loadConfig(conf);
  const auth = new AuthAPIService({
    jwtSecret: cfg.jwtSecret,
    jwtExpired: cfg.jwtExpired,
  });
  const requireJWT = jwt(cfg.jwtSecret);
  const { host, port } = parseAddr(cfg.addr);

  const server = http.createServer(async (req, res) => {
    if (req.method === "OPTIONS") {
      writeJSON(res, 204, {});
      return;
    }
    if (req.method === "GET" && req.url === "/healthz") {
      writeJSON(res, 200, { status: "ok" });
      return;
    }
    try {
      if (req.method === "POST" && req.url === "/v1/auth/sign-up") {
        const body = await readJSON(req);
        const result = await auth.signUp(body);
        writeJSON(res, result.status, result.body);
        return;
      }
      if (req.method === "POST" && req.url === "/v1/auth/sign-in") {
        const body = await readJSON(req);
        const result = await auth.signIn(body);
        writeJSON(res, result.status, result.body);
        return;
      }
      if (req.method === "GET" && req.url === "/v1/auth/me") {
        requireJWT(req);
        const claims = extractClaimsFromContext(req.ctx || {});
        writeJSON(res, 200, { user: claims });
        return;
      }
      writeJSON(res, 404, { code: "NOT_FOUND", message: "route not found" });
    } catch (err) {
      const isAuthError =
        err.message === "missing authorization header" ||
        err.message === "invalid token" ||
        err.message === "invalid token signature" ||
        err.message === "token expired" ||
        err.message === "no user claims in context";
      writeJSON(res, isAuthError ? 401 : 400, {
        code: isAuthError ? "UNAUTHORIZED" : "BAD_REQUEST",
        message: err.message,
      });
    }
  });

  server.listen(port, host, () => {
    // eslint-disable-next-line no-console
    console.log(`auth service listening on ${host}:${port}`);
  });
}

main().catch((err) => {
  // eslint-disable-next-line no-console
  console.error(err);
  process.exit(1);
});
