const crypto = require("crypto");
const { promisify } = require("util");
const { Struct } = require("google-protobuf/google/protobuf/struct_pb.js");
const authpb = require("../../api/auth/v1/auth_pb.js");

const scryptAsync = promisify(crypto.scrypt);

function b64url(s) {
  return Buffer.from(s)
    .toString("base64")
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_");
}

function signToken(payload, secret) {
  const header = b64url(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const body = b64url(JSON.stringify(payload));
  const content = `${header}.${body}`;
  const sig = crypto
    .createHmac("sha256", secret)
    .update(content)
    .digest("base64")
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_");
  return `${content}.${sig}`;
}

class AuthAPIService {
  constructor(options = {}) {
    this.jwtSecret = options.jwtSecret || "dev-secret";
    this.jwtExpired = Number(options.jwtExpired || 3600);
    this.nextID = 1;
    this.users = new Map(); // email -> user
  }

  async signUp(input) {
    const req = new authpb.SignUpRequest();
    req.setEmail(String(input?.email || "").trim().toLowerCase());
    req.setPassword(String(input?.password || ""));
    req.setName(String(input?.name || "").trim());

    if (!req.getEmail() || !req.getPassword() || !req.getName()) {
      return this.err(400, "INVALID_ARGUMENT", "email, password, name are required");
    }
    if (this.users.has(req.getEmail())) {
      return this.err(409, "ALREADY_EXISTS", "email already exists");
    }

    const salt = crypto.randomBytes(16).toString("hex");
    const hash = await scryptAsync(req.getPassword(), salt, 64);
    const user = {
      id: this.nextID++,
      email: req.getEmail(),
      name: req.getName(),
      passwordHash: `${salt}:${Buffer.from(hash).toString("hex")}`,
    };
    this.users.set(user.email, user);

    const reply = new authpb.SignUpReply();
    reply.setUserId(user.id);
    return { status: 200, body: { user_id: reply.getUserId() } };
  }

  async signIn(input) {
    const req = new authpb.SignInRequest();
    req.setEmail(String(input?.email || "").trim().toLowerCase());
    req.setPassword(String(input?.password || ""));

    const user = this.users.get(req.getEmail());
    if (!user) {
      return this.err(401, "UNAUTHORIZED", "invalid email or password");
    }

    const [salt, saved] = user.passwordHash.split(":");
    const hash = await scryptAsync(req.getPassword(), salt, 64);
    const ok = crypto.timingSafeEqual(Buffer.from(saved, "hex"), Buffer.from(hash));
    if (!ok) {
      return this.err(401, "UNAUTHORIZED", "invalid email or password");
    }

    const now = Math.floor(Date.now() / 1000);
    const token = signToken(
      {
        sub: String(user.id),
        user_id: user.id,
        email: user.email,
        name: user.name,
        is_admin: false,
        roles: ["user"],
        iat: now,
        exp: now + this.jwtExpired,
      },
      this.jwtSecret,
    );

    const reply = new authpb.SignInReply();
    reply.setAccessToken(token);
    reply.setTokenType("Bearer");
    reply.setExpiresIn(this.jwtExpired);
    reply.setUser(
      Struct.fromJavaScript({
        id: user.id,
        email: user.email,
        name: user.name,
      }),
    );

    return {
      status: 200,
      body: {
        access_token: reply.getAccessToken(),
        token_type: reply.getTokenType(),
        expires_in: reply.getExpiresIn(),
        user: reply.getUser()?.toJavaScript() || null,
      },
    };
  }

  err(status, code, message) {
    return { status, body: { code, message } };
  }
}

module.exports = { AuthAPIService };
