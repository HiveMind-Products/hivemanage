import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Link, Outlet, useLocation, useParams } from "react-router";
import { AppSidebar } from "./AppSidebar";
import { Separator } from "@/components/ui/separator";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { ModeToggle } from "@/components/theme/ModeToggle";
import { Params } from "@/typings/router";
import { useCurrentOrganization } from "@/features/organizations/api/useCurrentOrganization";
import { Suspense } from "react";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import { Button } from "@/components/ui/button";

const titleCase = (s: string) =>
  s.charAt(0).toUpperCase() + s.slice(1).replace(/-/g, " ");

export function AppLayout() {
  const location = useLocation();
  const params = useParams<Params>();
  const organization = useCurrentOrganization(params.organizationId);

  const paths = location.pathname.split("/").filter((p) => p);
  const crumbs = paths.slice(2);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-20 flex h-12 shrink-0 items-center gap-2 border-b border-border/70 bg-background/80 px-4 backdrop-blur-md supports-[backdrop-filter]:bg-background/60">
          <nav className="flex min-w-0 items-center gap-2">
            <SidebarTrigger className="text-muted-foreground hover:text-foreground" />
            <Separator orientation="vertical" className="mr-1 h-4" />
            <Breadcrumb>
              <BreadcrumbList className="gap-1.5 sm:gap-1.5">
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink
                    asChild
                    className="text-sm font-medium text-foreground transition-colors hover:text-foreground/80"
                  >
                    <Link to={`/app/${params.organizationId}`}>
                      {organization.data?.name ?? "Dashboard"}
                    </Link>
                  </BreadcrumbLink>
                </BreadcrumbItem>
                {crumbs.map((path, i) => (
                  <span key={path} className="flex items-center gap-1.5">
                    <BreadcrumbSeparator className="hidden text-muted-foreground/50 md:block" />
                    <BreadcrumbItem>
                      {i === crumbs.length - 1 ? (
                        <BreadcrumbPage className="text-sm font-medium">
                          {titleCase(path)}
                        </BreadcrumbPage>
                      ) : (
                        <span className="text-sm text-muted-foreground">
                          {titleCase(path)}
                        </span>
                      )}
                    </BreadcrumbItem>
                  </span>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
          </nav>
          <div className="flex-1" />
          <div className="flex items-center gap-1.5">
            <ModeToggle />
          </div>
        </header>
        <main className="min-h-[calc(100svh-3rem)] p-4 sm:p-6">
          <ErrorBoundary
            fallback={(error, reset) => (
              <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 p-6 text-center">
                <div className="space-y-1">
                  <h2 className="text-xl font-semibold">
                    Couldn't load this page
                  </h2>
                  <p className="max-w-md text-sm text-muted-foreground">
                    {error.message || "An unexpected error occurred."}
                  </p>
                </div>
                <Button variant="outline" onClick={reset}>
                  Try again
                </Button>
              </div>
            )}
          >
            <Suspense
              fallback={
                <div className="flex min-h-[50vh] items-center justify-center text-sm text-muted-foreground">
                  Loading…
                </div>
              }
            >
              <Outlet />
            </Suspense>
          </ErrorBoundary>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
