import { useParams } from "react-router";
import { useSession } from "../api/useSession";
import type { PermissionAction, PermissionModule } from "@/typings/permissions";

export function usePermission(module: PermissionModule, action: PermissionAction = "read") {
  const { organizationId } = useParams<{ organizationId: string }>();
  const { data: session } = useSession();

  if (session?.isAdmin) return true;
  if (!organizationId) return false;

  const access = session?.permissionsByOrganization?.[organizationId]?.[module];
  return action === "write" ? !!access?.write : !!access?.read;
}

export function firstReadablePath(permissions: Record<string, { read: boolean }> | undefined) {
  if (permissions?.overview?.read) return "";
  if (permissions?.storage?.read) return "storage";
  if (permissions?.logs?.read) return "logs";
  if (permissions?.tokens?.read) return "tokens";
  if (permissions?.team?.read) return "team";
  return "";
}
