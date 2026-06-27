import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { usePermission } from "@/features/auth/hooks/use-permission";
import { QueryKeys } from "@/typings/query";
import { permissionModules, presetPermissions, rolePresets } from "@/typings/permissions";
import type { MemberPermissions, MemberRole, PermissionModule } from "@/typings/permissions";
import type {
  CreateMemberRequest,
  CreateMemberResponse,
  OrganizationMember,
  UpdateMemberRequest,
} from "@/typings/member";
import type { CreateInviteRequest, InviteResponse } from "@/typings/invite";
import { fetchApi } from "@/utils/http-util";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Copy, Save, Trash2, UserPlus } from "lucide-react";
import { useMemo, useState } from "react";
import { useParams } from "react-router";
import { toast } from "sonner";

const moduleLabels: Record<PermissionModule, string> = {
  overview: "Overview",
  logs: "Logs",
  storage: "Storage",
  tokens: "Tokens",
  team: "Team",
};

const defaultRole: MemberRole = "VIEWER";

export function OrganizationTeamRoute() {
  const { organizationId } = useParams<{ organizationId: string }>();
  const canWrite = usePermission("team", "write");
  const queryClient = useQueryClient();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<MemberRole>(defaultRole);
  const [permissions, setPermissions] = useState<MemberPermissions>(() => presetPermissions(defaultRole));
  const [drafts, setDrafts] = useState<Record<number, { role: MemberRole; permissions: MemberPermissions }>>({});

  const [inviteDiscord, setInviteDiscord] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<MemberRole>(defaultRole);
  const [invitePermissions, setInvitePermissions] = useState<MemberPermissions>(() => presetPermissions(defaultRole));

  const membersQuery = useQuery({
    queryKey: [QueryKeys.Members, organizationId],
    enabled: !!organizationId,
    queryFn: () => fetchApi<OrganizationMember[]>("/api/dash/organization/" + organizationId + "/member"),
  });

  const invitesQuery = useQuery({
    queryKey: [QueryKeys.Invites, organizationId],
    enabled: !!organizationId && canWrite,
    queryFn: () => fetchApi<InviteResponse[]>("/api/dash/organization/" + organizationId + "/invite"),
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: [QueryKeys.Members, organizationId] });
    queryClient.invalidateQueries({ queryKey: [QueryKeys.Session] });
    queryClient.invalidateQueries({ queryKey: ["organizations"] });
  };

  const createMember = useMutation({
    mutationFn: (body: CreateMemberRequest) =>
      fetchApi<CreateMemberResponse>("/api/dash/organization/" + organizationId + "/member", {
        method: "POST",
        body: JSON.stringify(body),
      }),
    onSuccess: (data) => {
      invalidate();
      setUsername("");
      setEmail("");
      setRole(defaultRole);
      setPermissions(presetPermissions(defaultRole));
      toast.success("Member created", { description: data ? "Password: " + data.password : undefined });
    },
  });

  const updateMember = useMutation({
    mutationFn: ({ memberId, body }: { memberId: number; body: UpdateMemberRequest }) =>
      fetchApi<OrganizationMember>("/api/dash/organization/" + organizationId + "/member/" + memberId, {
        method: "PATCH",
        body: JSON.stringify(body),
      }),
    onSuccess: (_, vars) => {
      invalidate();
      setDrafts((current) => {
        const next = { ...current };
        delete next[vars.memberId];
        return next;
      });
    },
  });

  const removeMember = useMutation({
    mutationFn: (memberId: number) =>
      fetchApi("/api/dash/organization/" + organizationId + "/member/" + memberId, { method: "DELETE" }),
    onSuccess: () => invalidate(),
  });

  const createInvite = useMutation({
    mutationFn: (body: CreateInviteRequest) =>
      fetchApi<InviteResponse>("/api/dash/organization/" + organizationId + "/invite", {
        method: "POST",
        body: JSON.stringify(body),
      }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.Invites, organizationId] });
      setInviteDiscord("");
      setInviteEmail("");
      setInviteRole(defaultRole);
      setInvitePermissions(presetPermissions(defaultRole));
      const link = data ? inviteLink(data) : undefined;
      if (link) {
        void copyToClipboard(link);
        toast.success("Invite created", { description: "Link copied to clipboard" });
      } else {
        toast.success("Invite created");
      }
    },
  });

  const members = membersQuery.data ?? [];

  function onRoleChange(nextRole: MemberRole) {
    setRole(nextRole);
    setPermissions(presetPermissions(nextRole));
  }

  function onCreate() {
    createMember.mutate({ username, email, role, permissions });
  }

  function onInviteRoleChange(nextRole: MemberRole) {
    setInviteRole(nextRole);
    setInvitePermissions(presetPermissions(nextRole));
  }

  function onCreateInvite() {
    createInvite.mutate({
      role: inviteRole,
      permissions: invitePermissions,
      discordUsername: inviteDiscord,
      email: inviteEmail,
    });
  }

  function draftFor(member: OrganizationMember) {
    return drafts[member.id] ?? { role: member.role, permissions: member.permissions };
  }

  function updateDraft(member: OrganizationMember, next: { role?: MemberRole; permissions?: MemberPermissions }) {
    setDrafts((current) => {
      const existing = draftFor(member);
      return {
        ...current,
        [member.id]: {
          role: next.role ?? existing.role,
          permissions: next.permissions ?? existing.permissions,
        },
      };
    });
  }

  const createDisabled = !canWrite || !username || !email || createMember.isPending;
  const inviteDisabled = !canWrite || (!inviteDiscord && !inviteEmail) || createInvite.isPending;
  const invites = invitesQuery.data ?? [];

  return (
    <main className="container mx-auto max-w-6xl space-y-6 py-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Team</h1>
        <p className="text-muted-foreground">Manage dashboard access for organization members.</p>
      </div>

      {canWrite && (
        <section className="space-y-4 rounded-md border p-4">
          <div className="grid gap-3 md:grid-cols-[1fr_1fr_180px_auto]">
            <Input placeholder="Username" value={username} onChange={(event) => setUsername(event.target.value)} />
            <Input placeholder="Email" value={email} onChange={(event) => setEmail(event.target.value)} />
            <RoleSelect value={role} onChange={onRoleChange} />
            <Button onClick={onCreate} disabled={createDisabled}>
              <UserPlus className="h-4 w-4" />
              Add member
            </Button>
          </div>
          <PermissionMatrix permissions={permissions} onChange={setPermissions} disabled={!canWrite} />
        </section>
      )}

      {canWrite && (
        <section className="space-y-4 rounded-md border p-4">
          <div>
            <h2 className="text-lg font-semibold">Invite via Discord</h2>
            <p className="text-sm text-muted-foreground">
              Create a shareable link. The invitee signs in with Discord to join with the role below.
            </p>
          </div>
          <div className="grid gap-3 md:grid-cols-[1fr_1fr_180px_auto]">
            <Input
              placeholder="Discord username (optional)"
              value={inviteDiscord}
              onChange={(event) => setInviteDiscord(event.target.value)}
            />
            <Input
              placeholder="Email (optional)"
              value={inviteEmail}
              onChange={(event) => setInviteEmail(event.target.value)}
            />
            <RoleSelect value={inviteRole} onChange={onInviteRoleChange} />
            <Button onClick={onCreateInvite} disabled={inviteDisabled}>
              <UserPlus className="h-4 w-4" />
              Create invite
            </Button>
          </div>
          <PermissionMatrix permissions={invitePermissions} onChange={setInvitePermissions} disabled={!canWrite} />

          {invites.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-medium">Pending invites</h3>
              <div className="divide-y rounded-md border">
                {invites.map((invite) => (
                  <div key={invite.id} className="flex items-center justify-between gap-3 p-3">
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium">
                        {invite.discordUsername || invite.email || "Anyone with the link"}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {invite.role}
                        {invite.expiresAt
                          ? " · expires " + new Date(invite.expiresAt).toLocaleDateString()
                          : ""}
                      </div>
                    </div>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        void copyToClipboard(inviteLink(invite));
                        toast.success("Invite link copied");
                      }}
                    >
                      <Copy className="h-4 w-4" />
                      Copy link
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          )}
        </section>
      )}

      <section className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Member</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Permissions</TableHead>
              {canWrite && <TableHead className="text-right">Actions</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {members.map((member) => {
              const draft = draftFor(member);
              return (
                <TableRow key={member.id}>
                  <TableCell>
                    <div className="font-medium">{member.name || member.email}</div>
                    <div className="text-xs text-muted-foreground">{member.email}</div>
                  </TableCell>
                  <TableCell>
                    {canWrite ? (
                      <RoleSelect
                        value={draft.role}
                        onChange={(nextRole) => updateDraft(member, { role: nextRole, permissions: presetPermissions(nextRole) })}
                      />
                    ) : (
                      draft.role
                    )}
                  </TableCell>
                  <TableCell>
                    <PermissionMatrix
                      permissions={draft.permissions}
                      disabled={!canWrite || draft.role === "ADMIN"}
                      compact
                      onChange={(nextPermissions) => updateDraft(member, { permissions: nextPermissions })}
                    />
                  </TableCell>
                  {canWrite && (
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          size="icon"
                          variant="outline"
                          onClick={() => updateMember.mutate({ memberId: member.id, body: draft })}
                          disabled={updateMember.isPending}
                        >
                          <Save className="h-4 w-4" />
                        </Button>
                        <Button
                          size="icon"
                          variant="destructive"
                          onClick={() => removeMember.mutate(member.id)}
                          disabled={removeMember.isPending}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </section>
    </main>
  );
}

// inviteLink turns the API's relative invite path into a full shareable URL.
function inviteLink(invite: InviteResponse): string {
  return invite.url.startsWith("http") ? invite.url : window.location.origin + invite.url;
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    // Clipboard may be unavailable (e.g. non-secure context); link is still shown in the list.
  }
}

function RoleSelect({ value, onChange }: { value: MemberRole; onChange: (role: MemberRole) => void }) {
  return (
    <Select value={value} onValueChange={(next) => onChange(next as MemberRole)}>
      <SelectTrigger className="w-full">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {rolePresets.map((role) => (
          <SelectItem key={role} value={role}>
            {role}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

function PermissionMatrix({
  permissions,
  onChange,
  disabled,
  compact = false,
}: {
  permissions: MemberPermissions;
  onChange: (permissions: MemberPermissions) => void;
  disabled: boolean;
  compact?: boolean;
}) {
  const rows = useMemo(() => permissionModules, []);

  function setAccess(module: PermissionModule, key: "read" | "write", value: boolean) {
    const next = { ...permissions, [module]: { ...permissions[module], [key]: value } };
    if (key === "write" && value) next[module].read = true;
    if (key === "read" && !value) next[module].write = false;
    onChange(next);
  }

  return (
    <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-5">
      {rows.map((module) => (
        <div key={module} className="rounded-md border p-2">
          <div className="mb-1.5 text-xs font-medium">{moduleLabels[module]}</div>
          <div className={compact ? "flex flex-wrap gap-x-3 gap-y-1" : "space-y-2"}>
            <label className="flex items-center gap-1.5 text-xs">
              <Checkbox
                checked={!!permissions[module]?.read}
                disabled={disabled}
                onCheckedChange={(checked) => setAccess(module, "read", checked === true)}
              />
              Read
            </label>
            <label className="flex items-center gap-1.5 text-xs">
              <Checkbox
                checked={!!permissions[module]?.write}
                disabled={disabled}
                onCheckedChange={(checked) => setAccess(module, "write", checked === true)}
              />
              Write
            </label>
          </div>
        </div>
      ))}
    </div>
  );
}
