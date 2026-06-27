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

export function SettingsRoute() {
  const canManageOrg = usePermission("team", "write");

  return (
    <div className="container mx-auto max-w-3xl space-y-6 py-2">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">Manage your account and organization.</p>
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
              className="bg-[#5865F2] text-white hover:bg-[#4752c4]"
              onClick={() => {
                window.location.href = "/api/dash/auth/discord/link";
              }}
            >
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
          {mismatch && <p className="text-xs text-red-600">Passwords do not match.</p>}
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
              <code className="rounded bg-muted px-2 py-1 text-sm">{organizationId}</code>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  void navigator.clipboard?.writeText(organizationId ?? "");
                  toast.success("Copied");
                }}
              >
                Copy
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
