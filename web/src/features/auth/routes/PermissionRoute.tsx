import type { ReactNode } from "react";
import type { PermissionModule } from "@/typings/permissions";
import { Navigate, Outlet, useParams } from "react-router";
import { useSession } from "../api/useSession";
import { firstReadablePath, usePermission } from "../hooks/use-permission";

function fallbackTarget(organizationId: string | undefined, fallback: string) {
  return "/app/" + organizationId + (fallback ? "/" + fallback : "");
}

export function PermissionRoute({ module, children }: { module: PermissionModule; children?: ReactNode }) {
  const { organizationId } = useParams<{ organizationId: string }>();
  const { data: session } = useSession();
  const canRead = usePermission(module, "read");

  if (canRead) return <>{children ?? <Outlet />}</>;

  const fallback = firstReadablePath(
    organizationId ? session?.permissionsByOrganization?.[organizationId] : undefined,
  );
  return <Navigate to={fallbackTarget(organizationId, fallback)} replace />;
}

export const PermissionElement = PermissionRoute;
