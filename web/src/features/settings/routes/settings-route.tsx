import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useSession } from "@/features/auth/api/useSession";
import { useCurrentOrganization } from "@/features/organizations/api/useCurrentOrganization";
import { usePermission } from "@/features/auth/hooks/use-permission";
import { QueryKeys } from "@/typings/query";
import { fetchApi } from "@/utils/http-util";
import { useCopyToClipboard } from "@/hooks/use-copy";
import { Check, Copy } from "lucide-react";

export function SettingsRoute() {
  const canManageOrg = usePermission("team", "write");

  return (
    <div className="mx-auto w-full max-w-3xl space-y-6">
      <div>
        <h1 className="text-lg font-semibold tracking-tight">Settings</h1>
        <p className="text-sm text-muted-foreground">
          Manage your account and organization.
        </p>
      </div>

      <Tabs defaultValue="account">
        <TabsList>
          <TabsTrigger value="account">Account</TabsTrigger>
          {canManageOrg && <TabsTrigger value="organization">Organization</TabsTrigger>}
        </TabsList>
        <TabsContent value="account" className="space-y-6 pt-4">
          <AccountSettings />
        </TabsContent>
        {canManageOrg && (
          <TabsContent value="organization" className="space-y-6 pt-4">
            <OrganizationSettings />
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}

function AccountSettings() {
  const { data: session } = useSession();
  const queryClient = useQueryClient();
  const [name, setName] = useState(session?.name ?? "");
  const [avatar, setAvatar] = useState(session?.avatar ?? "");

  const invalidateSession = () =>
    queryClient.invalidateQueries({ queryKey: [QueryKeys.Session] });

  const updateProfile = useMutation({
    mutationFn: (body: { name: string; avatar: string }) =>
      fetchApi("/api/dash/auth/me", { method: "PATCH", body: JSON.stringify(body) }),
    onSuccess: () => {
      invalidateSession();
      toast.success("Profile updated");
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : "Failed to update profile"),
  });

  const unlinkDiscord = useMutation({
    mutationFn: () => fetchApi("/api/dash/auth/discord/unlink", { method: "POST" }),
    onSuccess: () => {
      invalidateSession();
      toast.success("Discord unlinked");
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : "Failed to unlink Discord"),
  });

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Profile</CardTitle>
          <CardDescription>Your display name and avatar.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Display name</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="avatar">Avatar URL</Label>
            <Input id="avatar" value={avatar} onChange={(e) => setAvatar(e.target.value)} placeholder="https://…" />
          </div>
          <div className="text-sm text-muted-foreground">
            Username: <span className="font-medium text-foreground">{session?.username}</span>
            {session?.email ? <> · {session.email}</> : null}
          </div>
          <Button
            onClick={() => updateProfile.mutate({ name, avatar })}
            disabled={updateProfile.isPending}
          >
            Save profile
          </Button>
        </CardContent>
      </Card>

      <ChangePasswordCard />

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Discord</CardTitle>
          <CardDescription>
            {session?.discordLinked
              ? "Your account is linked to Discord."
              : "Link Discord to sign in with it."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {session?.discordLinked ? (
            <Button
              variant="outline"
              onClick={() => unlinkDiscord.mutate()}
              disabled={unlinkDiscord.isPending}
            >
              Unlink Discord
            </Button>
          ) : (
            <Button
              className="bg-discord text-discord-foreground hover:bg-discord/90"
              onClick={() => {
                window.location.href = "/api/dash/auth/discord/link";
              }}
            >
              <DiscordIcon className="size-4" />
              Link Discord
            </Button>
          )}
        </CardContent>
      </Card>
    </>
  );
}

function ChangePasswordCard() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");

  const changePassword = useMutation({
    mutationFn: (body: { currentPassword: string; newPassword: string }) =>
      fetchApi("/api/dash/auth/password", { method: "POST", body: JSON.stringify(body) }),
    onSuccess: () => {
      setCurrent("");
      setNext("");
      setConfirm("");
      toast.success("Password changed");
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : "Failed to change password"),
  });

  const mismatch = confirm.length > 0 && next !== confirm;
  const disabled = next.length < 8 || mismatch || changePassword.isPending;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Password</CardTitle>
        <CardDescription>At least 8 characters.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label htmlFor="current">Current password</Label>
          <Input id="current" type="password" value={current} onChange={(e) => setCurrent(e.target.value)} />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="new">New password</Label>
          <Input id="new" type="password" value={next} onChange={(e) => setNext(e.target.value)} />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="confirm">Confirm new password</Label>
          <Input id="confirm" type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} />
          {mismatch && <p className="text-xs text-destructive">Passwords do not match.</p>}
        </div>
        <Button
          onClick={() => changePassword.mutate({ currentPassword: current, newPassword: next })}
          disabled={disabled}
        >
          Change password
        </Button>
      </CardContent>
    </Card>
  );
}

function OrganizationSettings() {
  const { organizationId } = useParams<{ organizationId: string }>();
  const { data: organization } = useCurrentOrganization(organizationId);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { copied, copy } = useCopyToClipboard();
  const [name, setName] = useState(organization?.name ?? "");
  const [confirmName, setConfirmName] = useState("");

  const renameOrg = useMutation({
    mutationFn: (body: { name: string }) =>
      fetchApi("/api/dash/organization/" + organizationId, { method: "PATCH", body: JSON.stringify(body) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.Organization, organizationId] });
      queryClient.invalidateQueries({ queryKey: [QueryKeys.Organizations] });
      toast.success("Organization renamed");
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : "Failed to rename organization"),
  });

  const deleteOrg = useMutation({
    mutationFn: () => fetchApi("/api/dash/organization/" + organizationId, { method: "DELETE" }),
    onSuccess: () => {
      queryClient.clear();
      toast.success("Organization deleted");
      navigate("/app");
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : "Failed to delete organization"),
  });

  // Show the persisted name once it loads, before any edits.
  const displayName = name || organization?.name || "";

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">General</CardTitle>
          <CardDescription>Organization name and identifier.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="org-name">Name</Label>
            <Input id="org-name" value={displayName} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid gap-2">
            <Label>Organization ID</Label>
            <div className="flex items-center gap-2">
              <code className="flex-1 truncate rounded-md border bg-muted px-2.5 py-1.5 font-mono text-sm">
                {organizationId}
              </code>
              <Button
                size="icon"
                variant="outline"
                aria-label="Copy organization ID"
                onClick={() => copy(organizationId ?? "", "Organization ID copied")}
              >
                {copied ? (
                  <Check className="size-4 text-success" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>
          </div>
          <Button
            onClick={() => renameOrg.mutate({ name: displayName })}
            disabled={!displayName || displayName === organization?.name || renameOrg.isPending}
          >
            Save
          </Button>
        </CardContent>
      </Card>

      <Card className="border-destructive/50">
        <CardHeader>
          <CardTitle className="text-base text-destructive">Danger zone</CardTitle>
          <CardDescription>
            Deleting an organization permanently removes its members, tokens, datasets, and file records.
            This cannot be undone.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Dialog onOpenChange={() => setConfirmName("")}>
            <DialogTrigger asChild>
              <Button variant="destructive">Delete organization</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Delete {organization?.name}?</DialogTitle>
                <DialogDescription>
                  Type <span className="font-semibold">{organization?.name}</span> to confirm. This permanently
                  deletes the organization and its data.
                </DialogDescription>
              </DialogHeader>
              <Input
                value={confirmName}
                onChange={(e) => setConfirmName(e.target.value)}
                placeholder="Organization name"
              />
              <DialogFooter>
                <DialogClose asChild>
                  <Button variant="outline">Cancel</Button>
                </DialogClose>
                <Button
                  variant="destructive"
                  onClick={() => deleteOrg.mutate()}
                  disabled={confirmName !== organization?.name || deleteOrg.isPending}
                >
                  Permanently delete
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </>
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
