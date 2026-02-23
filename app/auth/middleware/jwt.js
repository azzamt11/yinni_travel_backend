const crypto = require("crypto");

function b64urlDecode(input) {
  const padded = input.replace(/-/g, "+").replace(/_/g, "/") + "===".slice((input.length + 3) % 4);
  return Buffer.from(padded, "base64").toString("utf8");
}

function verifyHS256(token, secret) {
  const parts = String(token || "").split(".");
  if (parts.length !== 3) {
    throw new Error("invalid token");
  }
  const [h, p, s] = parts;
  const expected = crypto
    .createHmac("sha256", secret)
    .update(`${h}.${p}`)
    .digest("base64")
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_");
  if (expected !== s) {
    throw new Error("invalid token signature");
  }
  const payload = JSON.parse(b64urlDecode(p));
  if (payload.exp && Math.floor(Date.now() / 1000) >= Number(payload.exp)) {
    throw new Error("token expired");
  }
  return payload;
}

function jwt(secret) {
  return (req) => {
    const auth = req.headers.authorization || "";
    if (!auth.startsWith("Bearer ")) {
      throw new Error("missing authorization header");
    }
    const token = auth.slice("Bearer ".length);
    const payload = verifyHS256(token, secret);

    req.ctx = req.ctx || {};
    req.ctx.user_id = Number(payload.user_id || payload.sub || 0);
    req.ctx.jwt_claims = {
      user_id: Number(payload.user_id || payload.sub || 0),
      email: payload.email || "",
      is_admin: Boolean(payload.is_admin),
      roles: Array.isArray(payload.roles) ? payload.roles.map(String) : [],
    };
  };
}

module.exports = { jwt };
