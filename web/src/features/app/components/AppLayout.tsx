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

export function AppLayout() {
  const location = useLocation();
  const params = useParams<Params>();
  const organization = useCurrentOrganization(params.organizationId);

  const paths = location.pathname.split("/").filter((p) => p);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex sticky top-0 bg-background h-10 shrink-0 items-center gap-2 border-b px-4">
          <nav className="flex items-center gap-2">
            <SidebarTrigger />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink asChild className="text-sm">
                    <Link to={`/app/${params.organizationId}`}>
                      {organization.data && organization.data.name}
                    </Link>
                  </BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator className="hidden md:block text-sm" />
                {paths.slice(2, paths.length).map((path) => (
                  <BreadcrumbItem key={path}>
                    <BreadcrumbPage>{path}</BreadcrumbPage>
                  </BreadcrumbItem>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
          </nav>
          <div className="flex-1" />
          <div className="flex items-center gap-4">
            <ModeToggle />
          </div>
        </header>
        <main className="p-4">
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
