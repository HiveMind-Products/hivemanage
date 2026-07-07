import { z } from "zod";

// Scopes a token can be granted. An empty selection means full access, matching
// the backend (a token with no scopes is unrestricted, for backward compat).
export const TOKEN_SCOPES = [
  { value: "storage:write", label: "Storage — upload/delete files" },
  { value: "storage:read", label: "Storage — list/read files" },
  { value: "logs:write", label: "Logs — ingest logs" },
] as const;

export type TokenScope = (typeof TOKEN_SCOPES)[number]["value"];

export const tokenSchema = z.object({
  identifier: z.string().min(1, "Identifier is required"),
  type: z.string(),
  scopes: z.array(z.string()),
  // ISO-8601 instant, or undefined for "never expires".
  expiresAt: z.string().optional(),
});

export type TokenParams = z.infer<typeof tokenSchema>;

export interface TokenPreview {
  token: string;
}

export interface Token {
  id: number;
  type: string;
  identifier: string;
  scopes?: string[] | null;
  expiresAt?: string | null;
}
