import { Navigate } from "react-router";
import { AuthForm } from "../components/AuthForm";
import { useTheme } from "@/components/theme/useTheme";
import { useSession } from "../api/useSession";

export const AuthRoute: React.FC = () => {
  const { theme } = useTheme();
  const { data: session, isPending: sessionPending } = useSession();

  if (sessionPending) {
    return null;
  }

  if (session) {
    return <Navigate to="/app" replace />;
  }

  let logoSrc = "/logos/logo-white.png";
  if (theme === "dark") {
    logoSrc = "/logos/logo-black.png";
  }
  return (
    <main className="min-h-screen bg-gradient-to-br from-background via-gray-100 to-gray-200 dark:from-background dark:via-gray-900 dark:to-gray-950 flex items-center justify-center">
      <div className="flex w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-sm">
          <div className="flex flex-col items-center mb-8">
            <img
              src={logoSrc}
              alt="Hivemanage Logo"
              className="w-55 h-10 mb-4 drop-shadow-lg"
            />
            <h1 className="text-3xl font-bold mb-1 text-center">Sign in to Hivemanage Lite</h1>
          </div>
          <AuthForm />
        </div>
      </div>
    </main>
  );
};
