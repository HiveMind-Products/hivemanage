import {
  ChartArea,
  Folders,
  KeyRound,
  Layers,
  Settings,
  Users,
  type LucideIcon,
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarHeader,
  SidebarFooter,
} from "@/components/ui/sidebar";
import { NavLink, useParams } from "react-router";
import { OrganizationSwitcher } from "@/features/app/components/OrganizationSwitcher";
import { useSession } from "@/features/auth/api/useSession";
import { NavUser } from "@/features/app/components/NavUser";
import { $api } from "@/lib/api/client";
import type { PermissionModule } from "@/typings/permissions";

const items: Array<{
  title: string;
  url: string;
  icon: LucideIcon;
  comingSoon: boolean;
  module?: PermissionModule;
}> = [
  { title: "Storage", url: "storage", icon: Folders, comingSoon: false, module: "storage" },
  { title: "Tokens", url: "tokens", icon: KeyRound, comingSoon: false, module: "tokens" },
  { title: "Logs", url: "logs", icon: Layers, comingSoon: false, module: "logs" },
  { title: "Team", url: "team", icon: Users, comingSoon: false, module: "team" },
  { title: "Usage", url: "#", icon: ChartArea, comingSoon: true },
  { title: "Settings", url: "#", icon: Settings, comingSoon: true },
];

export function AppSidebar() {
  const session = useSession();
  const params = useParams<{ organizationId: string }>();
  const version = $api.useQuery("get", "/dash/system/version");
  const currentPermissions = params.organizationId
    ? session.data?.permissionsByOrganization?.[params.organizationId]
    : undefined;

  const visibleItems = items.filter((item) => {
    if (!item.module) return true;
    if (session.data?.isAdmin) return true;
    return !!currentPermissions?.[item.module]?.read;
  });

  return (
    <Sidebar className="h-full">
      <div className="flex flex-col h-full">
        <SidebarHeader>
          <OrganizationSwitcher />
        </SidebarHeader>
        <SidebarContent className="flex-1">
          <SidebarGroup>
            <SidebarGroupLabel className="flex items-center justify-between">
              <span>Hivemanage Lite ({version.data?.current || "..."})</span>
              {version.data?.update_available && (
                <span
                  className="flex h-2 w-2 rounded-full bg-blue-600"
                  title={"New version available: " + version.data.latest}
                />
              )}
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {visibleItems.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild>
                      <NavLink to={item.url}>
                        <item.icon />
                        <span>{item.title}</span>
                      </NavLink>
                    </SidebarMenuButton>
                    {item.comingSoon && <SidebarMenuBadge>Coming soon</SidebarMenuBadge>}
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          {session.data && (
            <NavUser
              user={{
                name: session.data.name || session.data.username,
                email: session.data.email || "",
                avatar: session.data.avatar || "",
                discordLinked: session.data.discordLinked ?? false,
              }}
            />
          )}
        </SidebarFooter>
      </div>
    </Sidebar>
  );
}
