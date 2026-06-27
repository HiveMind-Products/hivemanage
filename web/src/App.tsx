import { lazy } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import { toast } from "sonner";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { AuthRoute } from "./features/auth/routes/AuthRoute";
import { InviteRoute } from "./features/auth/routes/InviteRoute";
import { AppDashboard } from "./features/app/routes/AppDashboard";
import { AppLayout } from "./features/app/components/AppLayout";
import { ThemeProvider } from "./components/theme/ThemeProvider";
import { ProtectedRoute } from "./features/auth/routes/ProtectedRoute";
import { NewOrganizationRoute } from "./features/organizations/routes/NewOrganizationRoute";
import { OrganizationSelectRoute } from "./features/organizations/routes/OrganizationSelectRoute";
import { Toaster } from "sonner";
import { OrganizationTeamRoute } from "./features/organizations/routes/organization-team-route";
import { SettingsRoute } from "./features/settings/routes/settings-route";
import { PermissionElement, PermissionRoute } from "./features/auth/routes/PermissionRoute";
const UsageRoute = lazy(() => import("./features/usage/routes/usage-route"));
const StorageRoute = lazy(
  () => import("./features/files/routes/storage-route"),
);
const FileRoute = lazy(() => import("./features/files/routes/file-route"));
const TokensRoute = lazy(() => import("./features/tokens/routes/tokens-route"));
const DatasetRoute = lazy(() => import("./features/logs/routes/dataset-route"));
const LogsRoute = lazy(() => import("./features/logs/routes/logs-route"));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 30, // 30s before data is considered stale
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      // Generic fallback toast; individual mutations may override with a
      // more specific message via their own onError.
      onError: (error) => {
        toast.error(
          error instanceof Error ? error.message : "Something went wrong",
        );
      },
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <BrowserRouter>
          <Toaster richColors position="bottom-right" />
          <ErrorBoundary>
          <Routes>
            <Route path="*" element={<div>404</div>} />
            <Route path="/" element={<Navigate to="/app" />} />
            <Route path="/auth" element={<AuthRoute />} />
            <Route path="/invite/:token" element={<InviteRoute />} />
            <Route path="/app">
              <Route index element={<OrganizationSelectRoute />} />
              <Route element={<ProtectedRoute />}>
                <Route
                  path="new-organization"
                  element={<NewOrganizationRoute />}
                />
                <Route path=":organizationId" element={<AppLayout />}>
                  <Route index element={<PermissionElement module="overview"><AppDashboard /></PermissionElement>} />
                  <Route path="tokens" element={<PermissionElement module="tokens"><TokensRoute /></PermissionElement>} />
                  <Route path="storage" element={<PermissionRoute module="storage" />}>
                    <Route index element={<StorageRoute />} />
                    <Route path=":fileId" element={<FileRoute />} />
                  </Route>
                  <Route path="logs" element={<PermissionRoute module="logs" />}>
                    <Route index element={<DatasetRoute />} />
                    <Route path=":datasetId" element={<LogsRoute />} />
                  </Route>
                  <Route path="team" element={<PermissionElement module="team"><OrganizationTeamRoute /></PermissionElement>} />
                  <Route path="usage" element={<PermissionElement module="overview"><UsageRoute /></PermissionElement>} />
                  <Route path="settings" element={<SettingsRoute />} />
                </Route>
              </Route>
            </Route>
          </Routes>
          </ErrorBoundary>
        </BrowserRouter>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;
