import type { MemberPermissions, MemberRole } from "./permissions";

export interface OrganizationMember {
  id: number;
  role: MemberRole;
  email: string;
  name: string;
  permissions: MemberPermissions;
}

export interface CreateMemberRequest {
  username: string;
  email: string;
  role: MemberRole;
  permissions: MemberPermissions;
}

export interface CreateMemberResponse {
  username: string;
  password: string;
}

export interface UpdateMemberRequest {
  role: MemberRole;
  permissions: MemberPermissions;
}
