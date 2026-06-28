import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useForm } from "react-hook-form";
import z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useLogin } from "../api/useLogin";
import { loginSchema, LoginSchema } from "@/typings/auth";

export function AuthForm({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
  const formMethods = useForm<LoginSchema>({
    resolver: zodResolver(loginSchema),
  });

  const { mutate, isPending, error: loginError } = useLogin();

  const handleLoginSubmit = async (data: z.infer<typeof loginSchema>) => {
    mutate(data);
  };

  const errorMessage = loginError instanceof Error ? loginError.message : null;

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Welcome back</CardTitle>
          <CardDescription>
            Sign in with your username and password.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...formMethods}>
            <form onSubmit={formMethods.handleSubmit(handleLoginSubmit)}>
              <div className="flex flex-col gap-6">
                {errorMessage && (
                  <div
                    role="alert"
                    className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
                  >
                    {errorMessage}
                  </div>
                )}
                <div className="grid gap-2">
                  <FormField
                    control={formMethods.control}
                    name="username"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Username</FormLabel>
                        <FormControl>
                          <Input {...field} disabled={isPending} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <div className="grid gap-2">
                  <FormField
                    control={formMethods.control}
                    name="password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Password</FormLabel>
                        <FormControl>
                          <Input
                            type="password"
                            placeholder="*********"
                            {...field}
                            disabled={isPending}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <Button type="submit" className="w-full" disabled={isPending}>
                  {isPending ? "Signing in…" : "Sign in"}
                </Button>
              </div>
            </form>
          </Form>
          <div className="my-5 flex items-center gap-3">
            <div className="h-px flex-1 bg-border" />
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              or
            </span>
            <div className="h-px flex-1 bg-border" />
          </div>
          <Button
            type="button"
            className="w-full bg-discord text-discord-foreground hover:bg-discord/90"
            disabled={isPending}
            onClick={() => {
              window.location.href = "/api/dash/auth/discord";
            }}
          >
            <DiscordIcon className="size-4" />
            Continue with Discord
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

function DiscordIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      className={className}
    >
      <path d="M20.317 4.369A19.79 19.79 0 0 0 15.432 3a13.7 13.7 0 0 0-.617 1.27 18.27 18.27 0 0 0-5.487 0A13 13 0 0 0 8.71 3a19.7 19.7 0 0 0-4.885 1.37C.72 8.96-.13 13.44.29 17.85a19.9 19.9 0 0 0 6.06 3.06c.49-.67.92-1.38 1.29-2.12-.71-.27-1.39-.6-2.03-.99.17-.12.34-.25.5-.38a14.2 14.2 0 0 0 12.18 0c.16.13.33.26.5.38-.64.39-1.32.72-2.03.99.37.74.8 1.45 1.29 2.12a19.8 19.8 0 0 0 6.06-3.06c.5-5.12-.85-9.56-3.55-13.48ZM8.02 15.33c-1.18 0-2.15-1.08-2.15-2.41 0-1.33.95-2.42 2.15-2.42 1.2 0 2.17 1.09 2.15 2.42 0 1.33-.95 2.41-2.15 2.41Zm7.96 0c-1.18 0-2.15-1.08-2.15-2.41 0-1.33.95-2.42 2.15-2.42 1.2 0 2.17 1.09 2.15 2.42 0 1.33-.95 2.41-2.15 2.41Z" />
    </svg>
  );
}
