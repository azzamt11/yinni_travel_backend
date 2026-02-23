function extractClaimsFromContext(ctx = {}) {
  if (ctx.jwt_claims) return ctx.jwt_claims;
  if (ctx.user_id) return { user_id: Number(ctx.user_id) };
  throw new Error("no user claims in context");
}

module.exports = { extractClaimsFromContext };
