export const AUTH_METHOD = {
  TLS: "tls",
  OIDC: "oidc",
  UNIX: "unix",
  BEARER: "bearer",
} as const;

// Type derived from the constants
export type LXDAuthMethod = (typeof AUTH_METHOD)[keyof typeof AUTH_METHOD];
