export type PermissionModule = "overview" | "logs" | "storage" | "tokens" | "team";
export type PermissionAction = "read" | "write";

export interface PermissionAccess {
  read: boolean;
  write: boolean;
}

export type MemberPermissions = Record<PermissionModule, PermissionAccess>;

export const permissionModules: PermissionModule[] = [
  "overview",
  "logs",
  "storage",
  "tokens",
  "team",
];

export const rolePresets = ["ADMIN", "EDITOR", "VIEWER"] as const;
export type MemberRole = (typeof rolePresets)[number];

export function presetPermissions(role: MemberRole): MemberPermissions {
  if (role === "ADMIN") return all(true, true);
  if (role === "EDITOR") {
    return {
      overview: { read: true, write: false },
      logs: { read: true, write: true },
      storage: { read: true, write: true },
      tokens: { read: false, write: false },
      team: { read: false, write: false },
    };
  }
  return {
    overview: { read: true, write: false },
    logs: { read: true, write: false },
    storage: { read: true, write: false },
    tokens: { read: false, write: false },
    team: { read: false, write: false },
  };
}

function all(read: boolean, write: boolean): MemberPermissions {
  return {
    overview: { read, write },
    logs: { read, write },
    storage: { read, write },
    tokens: { read, write },
    team: { read, write },
  };
}
