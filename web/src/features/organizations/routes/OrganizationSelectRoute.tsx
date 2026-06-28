import { Navigate, NavLink, useNavigate, useSearchParams } from "react-router";
import { useOrganizations } from "../api/useOrganizations";
import { useSession, useLogout } from "@/features/auth/api/useSession";
import { CirclePlus, ChevronRight, LogOut } from "lucide-react";
import { useEffect } from "react";
import { toast } from "sonner";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Wordmark } from "@/components/brand/logo";

export function OrganizationSelectRoute() {
  const { data, isLoading } = useOrganizations();
  const session = useSession();
  const { logout } = useLogout();
  const orgCount = data?.length || 0;
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  useEffect(() => {
    const discord = searchParams.get("discord");
    if (!discord) return;
    if (discord === "linked") toast.success("Discord account linked");
    else if (discord === "taken") toast.error("That Discord account is already linked to another user");
    else if (discord === "error") toast.error("Failed to link Discord account");
    const next = new URLSearchParams(searchParams);
    next.delete("discord");
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  if (!session.data && !session.isPending) {
    return <Navigate to="/auth" />;
  }

  if (isLoading) {
    return (
      <div className="flex min-h-svh items-center justify-center bg-background text-sm text-muted-foreground">
        Loading…
      </div>
    );
  }

  if (!data || data.length === 0) {
    return <Navigate to="/app/new-organization" />;
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-background p-6">
      <Wordmark markClassName="size-7" className="text-base" />
      <Card className="flex w-full max-w-md flex-col">
        <CardHeader className="border-b">
          <CardTitle className="text-base">Select a workspace</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div className="max-h-72 overflow-y-auto px-3 py-3">
            <ul className="flex flex-col gap-1">
              {data.map((org) => (
                <li key={org.id} className="w-full">
                  <NavLink
                    to={`/app/${org.id}`}
                    state={{ organization: org }}
                    className="group flex items-center gap-3 rounded-md px-3 py-2.5 transition-colors hover:bg-accent"
                  >
                    <span className="flex size-9 items-center justify-center rounded-md bg-primary/10 text-sm font-semibold text-primary">
                      {org.name.charAt(0).toUpperCase()}
                    </span>
                    <div className="min-w-0 grow">
                      <p className="truncate text-sm font-medium">{org.name}</p>
                      <p className="truncate text-xs text-muted-foreground">
                        ID: {org.id}
                      </p>
                    </div>
                    <ChevronRight className="size-4 shrink-0 text-muted-foreground/50 transition-colors group-hover:text-muted-foreground" />
                  </NavLink>
                </li>
              ))}
            </ul>
          </div>
          <div className="border-t p-3">
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={() => navigate("/app/new-organization")}
            >
              <CirclePlus className="size-4" />
              Create new workspace
            </Button>
          </div>
        </CardContent>
        <CardFooter className="flex items-center justify-between border-t px-4 py-3">
          <span className="text-xs text-muted-foreground">
            {orgCount} workspace{orgCount === 1 ? "" : "s"}
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={logout}
            className="gap-1.5 text-muted-foreground hover:text-foreground"
          >
            <LogOut className="size-4" />
            Sign out
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
