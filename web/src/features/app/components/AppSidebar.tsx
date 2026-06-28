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
import { NavLink, useLocation, useParams } from "react-router";
import { OrganizationSwitcher } from "@/features/app/components/OrganizationSwitcher";
import { useSession } from "@/features/auth/api/useSession";
import { NavUser } from "@/features/app/components/NavUser";
import { Logo } from "@/components/brand/logo";
import { cn } from "@/lib/utils";
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
  { title: "Usage", url: "usage", icon: ChartArea, comingSoon: false, module: "overview" },
  { title: "Settings", url: "settings", icon: Settings, comingSoon: false },
];

export function AppSidebar() {
  const session = useSession();
  const location = useLocation();
  const params = useParams<{ organizationId: string }>();
  const version = $api.useQuery("get", "/dash/system/version");
  const currentPermissions = params.organizationId
    ? session.data?.permissionsByOrganization?.[params.organizationId]
    : undefined;

  const base = `/app/${params.organizationId}`;

  const visibleItems = items.filter((item) => {
    if (!item.module) return true;
    if (session.data?.isAdmin) return true;
    return !!currentPermissions?.[item.module]?.read;
  });

  return (
    <Sidebar className="h-full" collapsible="icon">
      <SidebarHeader className="gap-2">
        <div className="flex h-9 items-center gap-2 px-2 group-data-[collapsible=icon]:px-0 group-data-[collapsible=icon]:justify-center">
          <Logo className="size-6" />
          <span className="flex items-baseline gap-1.5 text-sm font-semibold tracking-tight group-data-[collapsible=icon]:hidden">
            hive<span className="-ml-1.5 font-medium text-muted-foreground">manage</span>
            {version.data?.current && (
              <span className="ml-0.5 rounded-full border px-1.5 py-px text-[10px] font-medium text-muted-foreground tabular-nums">
                {version.data.current}
              </span>
            )}
            {version.data?.update_available && (
              <span
                className="size-1.5 rounded-full bg-primary"
                title={"New version available: " + version.data.latest}
              />
            )}
          </span>
        </div>
        <OrganizationSwitcher />
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Platform</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {visibleItems.map((item) => {
                const href = `${base}/${item.url}`;
                const active =
                  location.pathname === href ||
                  location.pathname.startsWith(href + "/");
                return (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      asChild
                      isActive={active}
                      tooltip={item.title}
                      className={cn(
                        "relative transition-colors duration-150",
                        active &&
                          "text-foreground before:absolute before:left-0 before:top-1/2 before:h-4 before:w-0.5 before:-translate-y-1/2 before:rounded-full before:bg-primary [&>svg]:text-primary",
                      )}
                    >
                      <NavLink to={item.url}>
                        <item.icon />
                        <span>{item.title}</span>
                      </NavLink>
                    </SidebarMenuButton>
                    {item.comingSoon && (
                      <SidebarMenuBadge>Soon</SidebarMenuBadge>
                    )}
                  </SidebarMenuItem>
                );
              })}
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
    </Sidebar>
  );
}
