import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Avatar,
  AvatarFallback,
} from "@/components/ui/avatar";
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
import { EmptyState } from "@/components/ui/empty-state";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
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
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { usePermission } from "@/features/auth/hooks/use-permission";
import { useCopyToClipboard } from "@/hooks/use-copy";
import { QueryKeys } from "@/typings/query";
import {
  permissionModules,
  presetPermissions,
  rolePresets,
} from "@/typings/permissions";
import type {
  MemberPermissions,
  MemberRole,
  PermissionAccess,
  PermissionModule,
} from "@/typings/permissions";
import type {
  CreateMemberRequest,
  CreateMemberResponse,
  OrganizationMember,
  UpdateMemberRequest,
} from "@/typings/member";
import type { CreateInviteRequest, InviteResponse } from "@/typings/invite";
import { fetchApi } from "@/utils/http-util";
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
} from "@tanstack/react-query";
import {
  Check,
  ChevronDown,
  Copy,
  Eye,
  Link2,
  Pencil,
  Save,
  Trash2,
  UserPlus,
  Users,
} from "lucide-react";
import { useState } from "react";
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

const roleDescriptions: Record<MemberRole, string> = {
  ADMIN: "Full access to everything, including team management.",
  EDITOR: "Can view and edit logs and storage.",
  VIEWER: "Read-only access to logs and storage.",
};

export function OrganizationTeamRoute() {
  const { organizationId } = useParams<{ organizationId: string }>();
  const canWrite = usePermission("team", "write");
  const queryClient = useQueryClient();
  const [drafts, setDrafts] = useState<
    Record<number, { role: MemberRole; permissions: MemberPermissions }>
  >({});

  const membersQuery = useQuery({
    queryKey: [QueryKeys.Members, organizationId],
    enabled: !!organizationId,
    queryFn: () =>
      fetchApi<OrganizationMember[]>(
        "/api/dash/organization/" + organizationId + "/member",
      ),
  });

  const invitesQuery = useQuery({
    queryKey: [QueryKeys.Invites, organizationId],
    enabled: !!organizationId && canWrite,
    queryFn: () =>
      fetchApi<InviteResponse[]>(
        "/api/dash/organization/" + organizationId + "/invite",
      ),
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: [QueryKeys.Members, organizationId] });
    queryClient.invalidateQueries({ queryKey: [QueryKeys.Session] });
    queryClient.invalidateQueries({ queryKey: [QueryKeys.Organizations] });
  };

  const createMember = useMutation({
    mutationFn: (body: CreateMemberRequest) =>
      fetchApi<CreateMemberResponse>(
        "/api/dash/organization/" + organizationId + "/member",
        { method: "POST", body: JSON.stringify(body) },
      ),
    onSuccess: (data) => {
      invalidate();
      toast.success("Member created", {
        description: data ? "Password: " + data.password : undefined,
      });
    },
    onError: (err) =>
      toast.error(err instanceof Error ? err.message : "Failed to create member"),
  });

  const updateMember = useMutation({
    mutationFn: ({ memberId, body }: { memberId: number; body: UpdateMemberRequest }) =>
      fetchApi<OrganizationMember>(
        "/api/dash/organization/" + organizationId + "/member/" + memberId,
        { method: "PATCH", body: JSON.stringify(body) },
      ),
    onSuccess: (_, vars) => {
      invalidate();
      setDrafts((current) => {
        const next = { ...current };
        delete next[vars.memberId];
        return next;
      });
      toast.success("Member updated");
    },
    onError: (err) =>
      toast.error(err instanceof Error ? err.message : "Failed to update member"),
  });

  const removeMember = useMutation({
    mutationFn: (memberId: number) =>
      fetchApi("/api/dash/organization/" + organizationId + "/member/" + memberId, {
        method: "DELETE",
      }),
    onSuccess: () => {
      invalidate();
      toast.success("Member removed");
    },
    onError: (err) =>
      toast.error(err instanceof Error ? err.message : "Failed to remove member"),
  });

  const createInvite = useMutation({
    mutationFn: (body: CreateInviteRequest) =>
      fetchApi<InviteResponse>(
        "/api/dash/organization/" + organizationId + "/invite",
        { method: "POST", body: JSON.stringify(body) },
      ),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.Invites, organizationId] });
      const link = data ? inviteLink(data) : undefined;
      if (link) {
        void navigator.clipboard?.writeText(link).catch(() => {});
        toast.success("Invite created", { description: "Link copied to clipboard" });
      } else {
        toast.success("Invite created");
      }
    },
    onError: (err) =>
      toast.error(err instanceof Error ? err.message : "Failed to create invite"),
  });

  const deleteInvite = useMutation({
    mutationFn: (inviteId: string) =>
      fetchApi("/api/dash/organization/" + organizationId + "/invite/" + inviteId, {
        method: "DELETE",
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.Invites, organizationId] });
      toast.success("Invite deleted");
    },
    onError: (err) =>
      toast.error(err instanceof Error ? err.message : "Failed to delete invite"),
  });

  const members = membersQuery.data ?? [];
  const invites = invitesQuery.data ?? [];

  function draftFor(member: OrganizationMember) {
    return drafts[member.id] ?? { role: member.role, permissions: member.permissions };
  }

  function updateDraft(
    member: OrganizationMember,
    next: { role?: MemberRole; permissions?: MemberPermissions },
  ) {
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

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold tracking-tight">Team</h1>
          <p className="text-sm text-muted-foreground">
            Manage who can access this organization and what they can do.
          </p>
        </div>
        {canWrite && (
          <div className="flex items-center gap-2">
            <InviteDialog create={createInvite} />
            <AddMemberDialog create={createMember} />
          </div>
        )}
      </div>

      <div className="overflow-hidden rounded-lg border bg-card">
        {members.length === 0 ? (
          <EmptyState
            icon={Users}
            title="No members yet"
            description="Add a member or share an invite link to get your team on board."
            className="border-0"
          />
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b bg-muted/40 hover:bg-muted/40">
                <TableHead className="h-10 font-medium text-muted-foreground">
                  Member
                </TableHead>
                <TableHead className="h-10 w-[150px] font-medium text-muted-foreground">
                  Role
                </TableHead>
                <TableHead className="h-10 font-medium text-muted-foreground">
                  Access
                </TableHead>
                {canWrite && (
                  <TableHead className="h-10 w-[120px] text-right font-medium text-muted-foreground">
                    <span className="sr-only">Actions</span>
                  </TableHead>
                )}
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.map((member) => {
                const draft = draftFor(member);
                const dirty = member.id in drafts;
                return (
                  <TableRow key={member.id} className="hover:bg-accent/30">
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="size-8">
                          <AvatarFallback className="bg-primary/10 text-xs font-medium text-primary">
                            {initials(member.name || member.email)}
                          </AvatarFallback>
                        </Avatar>
                        <div className="min-w-0">
                          <div className="truncate text-sm font-medium">
                            {member.name || member.email}
                          </div>
                          <div className="truncate text-xs text-muted-foreground">
                            {member.email}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {canWrite ? (
                        <RoleSelect
                          value={draft.role}
                          onChange={(nextRole) =>
                            updateDraft(member, {
                              role: nextRole,
                              permissions: presetPermissions(nextRole),
                            })
                          }
                        />
                      ) : (
                        <span className="text-sm">{draft.role}</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2.5">
                        <AccessSummary permissions={draft.permissions} />
                        {canWrite && (
                          <EditAccessPopover
                            permissions={draft.permissions}
                            disabled={draft.role === "ADMIN"}
                            onChange={(nextPermissions) =>
                              updateDraft(member, { permissions: nextPermissions })
                            }
                          />
                        )}
                      </div>
                    </TableCell>
                    {canWrite && (
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1.5">
                          <Button
                            size="sm"
                            variant={dirty ? "default" : "outline"}
                            onClick={() =>
                              updateMember.mutate({ memberId: member.id, body: draft })
                            }
                            disabled={!dirty || updateMember.isPending}
                          >
                            <Save className="size-3.5" />
                            Save
                          </Button>
                          <RemoveMemberDialog
                            name={member.name || member.email}
                            onConfirm={() => removeMember.mutate(member.id)}
                            pending={removeMember.isPending}
                          />
                        </div>
                      </TableCell>
                    )}
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </div>

      {canWrite && invites.length > 0 && (
        <section className="space-y-2">
          <h2 className="text-sm font-medium text-muted-foreground">
            Pending invites
          </h2>
          <div className="divide-y overflow-hidden rounded-lg border bg-card">
            {invites.map((invite) => (
              <div
                key={invite.id}
                className="flex items-center justify-between gap-3 px-4 py-3"
              >
                <div className="min-w-0">
                  <div className="truncate text-sm font-medium">
                    {invite.discordUsername || invite.email || "Anyone with the link"}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {invite.role}
                    {invite.expiresAt
                      ? " · expires " +
                        new Date(invite.expiresAt).toLocaleDateString()
                      : ""}
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-1.5">
                  <CopyLinkButton url={inviteLink(invite)} />
                  <Button
                    size="icon"
                    variant="ghost"
                    className="text-muted-foreground hover:text-destructive"
                    onClick={() => deleteInvite.mutate(invite.id)}
                    disabled={deleteInvite.isPending}
                    aria-label="Delete invite"
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}

// inviteLink turns the API's relative invite path into a full shareable URL.
function inviteLink(invite: InviteResponse): string {
  return invite.url.startsWith("http")
    ? invite.url
    : window.location.origin + invite.url;
}

function initials(value: string): string {
  const parts = value.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

function RoleSelect({
  value,
  onChange,
  className,
}: {
  value: MemberRole;
  onChange: (role: MemberRole) => void;
  className?: string;
}) {
  return (
    <Select value={value} onValueChange={(next) => onChange(next as MemberRole)}>
      <SelectTrigger size="sm" className={className ?? "w-full"}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {rolePresets.map((role) => (
          <SelectItem key={role} value={role}>
            <span className="capitalize">{role.toLowerCase()}</span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

/**
 * Per-module access as two independently-selectable capabilities — a member can
 * hold View, Edit, or both at once. Edit implies View; clearing View clears Edit.
 */
function AccessToggles({
  access,
  onChange,
  disabled,
}: {
  access: PermissionAccess | undefined;
  onChange: (access: PermissionAccess) => void;
  disabled?: boolean;
}) {
  const prevRead = !!access?.read;
  const prevWrite = !!access?.write;
  const value: string[] = [];
  if (prevRead) value.push("read");
  if (prevWrite) value.push("write");

  function handleChange(next: string[]) {
    const nextRead = next.includes("read");
    const nextWrite = next.includes("write");
    let result: PermissionAccess;
    if (nextWrite && !prevWrite)
      result = { read: true, write: true }; // enable Edit → also View
    else if (!nextWrite && prevWrite)
      result = { read: true, write: false }; // disable Edit → keep View
    else if (nextRead && !prevRead)
      result = { read: true, write: false }; // enable View
    else if (!nextRead && prevRead)
      result = { read: false, write: false }; // disable View → clear all
    else result = { read: prevRead, write: prevWrite };
    onChange(result);
  }

  return (
    <ToggleGroup
      type="multiple"
      variant="outline"
      size="sm"
      value={value}
      onValueChange={handleChange}
      disabled={disabled}
    >
      <ToggleGroupItem
        value="read"
        aria-label="View"
        className="gap-1.5 px-3 data-[state=on]:bg-primary/15 data-[state=on]:text-primary"
      >
        <Eye className="size-3.5" />
        View
      </ToggleGroupItem>
      <ToggleGroupItem
        value="write"
        aria-label="Edit"
        className="gap-1.5 px-3 data-[state=on]:bg-primary/15 data-[state=on]:text-primary"
      >
        <Pencil className="size-3.5" />
        Edit
      </ToggleGroupItem>
    </ToggleGroup>
  );
}

function PermissionEditor({
  permissions,
  onChange,
  disabled,
}: {
  permissions: MemberPermissions;
  onChange: (permissions: MemberPermissions) => void;
  disabled?: boolean;
}) {
  return (
    <div className="divide-y overflow-hidden rounded-md border">
      {permissionModules.map((module) => (
        <div
          key={module}
          className="flex items-center justify-between gap-3 px-3 py-2.5"
        >
          <span className="text-sm font-medium">{moduleLabels[module]}</span>
          <AccessToggles
            access={permissions[module]}
            disabled={disabled}
            onChange={(access) => onChange({ ...permissions, [module]: access })}
          />
        </div>
      ))}
    </div>
  );
}

// Compact, human-readable summary of a member's access for the table.
function summarizeModules(modules: PermissionModule[]): string {
  const names = modules.map((m) => moduleLabels[m]);
  if (names.length <= 2) return names.join(", ");
  return `${names.slice(0, 2).join(", ")} +${names.length - 2}`;
}

function AccessSummary({ permissions }: { permissions: MemberPermissions }) {
  const write = permissionModules.filter((m) => permissions[m]?.write);
  const read = permissionModules.filter(
    (m) => permissions[m]?.read && !permissions[m]?.write,
  );

  let text: string;
  if (write.length === permissionModules.length) {
    text = "Full access";
  } else if (write.length === 0 && read.length === 0) {
    text = "No access";
  } else if (write.length === 0 && read.length === permissionModules.length) {
    text = "View only";
  } else {
    const parts: string[] = [];
    if (write.length) parts.push(`Edit ${summarizeModules(write)}`);
    if (read.length) parts.push(`View ${summarizeModules(read)}`);
    text = parts.join(" · ");
  }

  return <span className="truncate text-sm text-muted-foreground">{text}</span>;
}

function EditAccessPopover({
  permissions,
  onChange,
  disabled,
}: {
  permissions: MemberPermissions;
  onChange: (permissions: MemberPermissions) => void;
  disabled?: boolean;
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          disabled={disabled}
          className="ml-auto gap-1.5"
        >
          Edit access
          <ChevronDown className="size-3.5 text-muted-foreground" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80">
        <div className="space-y-2.5">
          <p className="text-sm font-medium">Module access</p>
          <PermissionEditor permissions={permissions} onChange={onChange} />
        </div>
      </PopoverContent>
    </Popover>
  );
}

function CopyLinkButton({ url }: { url: string }) {
  const { copied, copy } = useCopyToClipboard();
  return (
    <Button
      size="sm"
      variant="outline"
      className="gap-1.5"
      onClick={() => copy(url, "Invite link copied")}
    >
      {copied ? (
        <Check className="size-3.5 text-success" />
      ) : (
        <Copy className="size-3.5" />
      )}
      {copied ? "Copied" : "Copy link"}
    </Button>
  );
}

function AddMemberDialog({
  create,
}: {
  create: UseMutationResult<
    CreateMemberResponse | undefined,
    Error,
    CreateMemberRequest
  >;
}) {
  const [open, setOpen] = useState(false);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<MemberRole>(defaultRole);
  const [permissions, setPermissions] = useState<MemberPermissions>(() =>
    presetPermissions(defaultRole),
  );

  function reset() {
    setUsername("");
    setEmail("");
    setRole(defaultRole);
    setPermissions(presetPermissions(defaultRole));
  }

  async function submit() {
    try {
      await create.mutateAsync({ username, email, role, permissions });
      reset();
      setOpen(false);
    } catch {
      // error toast handled by the mutation; keep the dialog open
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) reset();
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <UserPlus className="size-4" />
          Add member
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add member</DialogTitle>
          <DialogDescription>
            Create an account for someone on your team and set their access.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 pt-1">
          <div className="grid gap-2">
            <Label htmlFor="add-username">Username</Label>
            <Input
              id="add-username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="jane"
              autoComplete="off"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="add-email">Email</Label>
            <Input
              id="add-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="jane@acme.io"
              autoComplete="off"
            />
          </div>
          <RoleField value={role} onChange={(r) => { setRole(r); setPermissions(presetPermissions(r)); }} />
          <AccessField
            permissions={permissions}
            onChange={setPermissions}
            disabled={role === "ADMIN"}
          />
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button
            onClick={submit}
            disabled={!username || !email || create.isPending}
          >
            {create.isPending ? "Adding…" : "Add member"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function InviteDialog({
  create,
}: {
  create: UseMutationResult<
    InviteResponse | undefined,
    Error,
    CreateInviteRequest
  >;
}) {
  const [open, setOpen] = useState(false);
  const [discordUsername, setDiscordUsername] = useState("");
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<MemberRole>(defaultRole);
  const [permissions, setPermissions] = useState<MemberPermissions>(() =>
    presetPermissions(defaultRole),
  );

  function reset() {
    setDiscordUsername("");
    setEmail("");
    setRole(defaultRole);
    setPermissions(presetPermissions(defaultRole));
  }

  async function submit() {
    try {
      await create.mutateAsync({ role, permissions, discordUsername, email });
      reset();
      setOpen(false);
    } catch {
      // error toast handled by the mutation
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) reset();
      }}
    >
      <DialogTrigger asChild>
        <Button variant="outline">
          <Link2 className="size-4" />
          Invite
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite via link</DialogTitle>
          <DialogDescription>
            Generate a shareable link. The invitee signs in with Discord to join
            with the access below. The link is copied to your clipboard.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 pt-1">
          <div className="grid gap-2">
            <Label htmlFor="invite-discord">Discord username</Label>
            <Input
              id="invite-discord"
              value={discordUsername}
              onChange={(e) => setDiscordUsername(e.target.value)}
              placeholder="Optional"
              autoComplete="off"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="invite-email">Email</Label>
            <Input
              id="invite-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Optional"
              autoComplete="off"
            />
          </div>
          <RoleField value={role} onChange={(r) => { setRole(r); setPermissions(presetPermissions(r)); }} />
          <AccessField
            permissions={permissions}
            onChange={setPermissions}
            disabled={role === "ADMIN"}
          />
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button onClick={submit} disabled={create.isPending}>
            {create.isPending ? "Creating…" : "Create invite"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RoleField({
  value,
  onChange,
}: {
  value: MemberRole;
  onChange: (role: MemberRole) => void;
}) {
  return (
    <div className="grid gap-2">
      <Label>Role</Label>
      <RoleSelect value={value} onChange={onChange} />
      <p className="text-xs text-muted-foreground">{roleDescriptions[value]}</p>
    </div>
  );
}

function AccessField({
  permissions,
  onChange,
  disabled,
}: {
  permissions: MemberPermissions;
  onChange: (permissions: MemberPermissions) => void;
  disabled?: boolean;
}) {
  return (
    <div className="grid gap-2">
      <Label>Access</Label>
      <PermissionEditor
        permissions={permissions}
        onChange={onChange}
        disabled={disabled}
      />
      {disabled && (
        <p className="text-xs text-muted-foreground">
          Admins always have full access.
        </p>
      )}
    </div>
  );
}

function RemoveMemberDialog({
  name,
  onConfirm,
  pending,
}: {
  name: string;
  onConfirm: () => void;
  pending: boolean;
}) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          size="icon"
          variant="ghost"
          aria-label={`Remove ${name}`}
          className="text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="size-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Remove {name}?</DialogTitle>
          <DialogDescription>
            They'll immediately lose access to this organization. This can't be
            undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <DialogClose asChild>
            <Button variant="destructive" onClick={onConfirm} disabled={pending}>
              {pending ? "Removing…" : "Remove member"}
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
