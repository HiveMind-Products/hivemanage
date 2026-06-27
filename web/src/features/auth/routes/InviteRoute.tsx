import { useEffect } from "react";
import { useParams, useSearchParams } from "react-router";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export const InviteRoute: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const [searchParams] = useSearchParams();
  const wrongDiscord = searchParams.get("error") === "wrong_discord";

  useEffect(() => {
    if (!token || wrongDiscord) {
      return;
    }
    // Persist the invite token so the Discord callback can redeem it after login.
    document.cookie = `fmlite_invite=${encodeURIComponent(token)}; path=/; max-age=600; samesite=lax`;
    window.location.href = "/api/dash/auth/discord";
  }, [token, wrongDiscord]);

  return (
    <main className="min-h-screen bg-gradient-to-br from-background via-gray-100 to-gray-200 dark:from-background dark:via-gray-900 dark:to-gray-950 flex items-center justify-center">
      <div className="w-full max-w-sm p-6">
        <Card>
          <CardHeader>
            <CardTitle>{wrongDiscord ? "Wrong Discord account" : "Joining organization…"}</CardTitle>
            <CardDescription>
              {wrongDiscord
                ? "This invite is for a different Discord account. Log out of Discord and try the link again with the invited account."
                : "Redirecting you to Discord to accept your invite."}
            </CardDescription>
          </CardHeader>
          {wrongDiscord && (
            <CardContent>
              <Button
                className="w-full"
                onClick={() => {
                  window.location.href = `/invite/${token}`;
                }}
              >
                Try again
              </Button>
            </CardContent>
          )}
        </Card>
      </div>
    </main>
  );
};
