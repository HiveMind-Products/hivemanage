//{"data":{"id":15,"name":"Admin","username":"admin","email":null,"avatar":null,"isAdmin":true},"status":"ok"}
import type { MemberPermissions } from "./permissions";

export interface Session {
  id: number;
  name?: string;
  username: string;
  email?: string;
  avatar?: string;
  isAdmin: boolean;
  discordLinked?: boolean;
  permissionsByOrganization?: Record<string, MemberPermissions>;
}
