import { Navigate } from "react-router";
import { AuthForm } from "../components/AuthForm";
import { Logo } from "@/components/brand/logo";
import { useSession } from "../api/useSession";

export const AuthRoute: React.FC = () => {
  const { data: session, isPending: sessionPending } = useSession();

  if (sessionPending) {
    return null;
  }

  if (session) {
    return <Navigate to="/app" replace />;
  }

  return (
    <main className="relative flex min-h-svh items-center justify-center overflow-hidden bg-background p-6">
      {/* Single, quiet brand glow — not a gradient wash */}
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-[-10%] h-[420px] w-[640px] -translate-x-1/2 rounded-full opacity-[0.07] blur-3xl"
        style={{ background: "var(--primary)" }}
      />
      <div className="relative w-full max-w-sm">
        <div className="mb-8 flex flex-col items-center text-center">
          <Logo className="mb-4 size-10" />
          <h1 className="text-2xl font-semibold tracking-tight">
            Sign in to hive<span className="text-muted-foreground">manage</span>
          </h1>
          <p className="mt-1.5 text-sm text-muted-foreground">
            Your self-hosted CDN and observability dashboard.
          </p>
        </div>
        <AuthForm />
        <p className="mt-6 text-center text-xs text-muted-foreground">
          hivemanage, by HiveMind
        </p>
      </div>
    </main>
  );
};
