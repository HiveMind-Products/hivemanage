import type { MemberPermissions, MemberRole } from "./permissions";

export interface CreateInviteRequest {
  role: MemberRole;
  permissions: MemberPermissions;
  discordUsername: string;
  email: string;
}

export interface InviteResponse {
  id: string;
  url: string;
  organizationId: string;
  role: MemberRole;
  discordUsername: string;
  email: string;
  expiresAt: string | null;
  acceptedAt: string | null;
  createdAt: string;
}
