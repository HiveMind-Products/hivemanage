import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useTokens } from "../api/useTokens";
import { DeleteTokenDialog } from "./DeleteTokenDialog";
import { EmptyState } from "@/components/ui/empty-state";
import { CreateTokenDialog } from "./CreateTokenDialog";
import { Badge } from "@/components/ui/badge";
import { KeyRound } from "lucide-react";

function formatExpiry(expiresAt?: string | null): string {
  if (!expiresAt) return "Never";
  const date = new Date(expiresAt);
  if (Number.isNaN(date.getTime())) return "Never";
  const expired = date.getTime() < Date.now();
  return `${expired ? "Expired " : ""}${date.toLocaleDateString()}`;
}

interface TokensTableProps {
  organizationId: string;
  canWrite: boolean;
}

export function TokensTable({ organizationId, canWrite }: TokensTableProps) {
  const tokens = useTokens(organizationId);

  if (!tokens || tokens.length === 0) {
    return (
      <EmptyState
        icon={KeyRound}
        title="No tokens yet"
        description="API tokens let your servers and scripts authenticate with the platform. Create one to get started."
        action={canWrite ? <CreateTokenDialog /> : undefined}
      />
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border bg-card">
      <Table>
        <TableHeader>
          <TableRow className="border-b bg-muted/40 hover:bg-muted/40">
            <TableHead className="h-10 font-medium text-muted-foreground">
              Identifier
            </TableHead>
            <TableHead className="h-10 font-medium text-muted-foreground">
              Scopes
            </TableHead>
            <TableHead className="h-10 font-medium text-muted-foreground">
              Expires
            </TableHead>
            {canWrite && (
              <TableHead className="h-10 w-16 text-right font-medium text-muted-foreground">
                <span className="sr-only">Actions</span>
              </TableHead>
            )}
          </TableRow>
        </TableHeader>
        <TableBody>
          {tokens.map((token) => (
            <TableRow key={token.id} className="transition-colors hover:bg-accent/40">
              <TableCell className="font-medium">{token.identifier}</TableCell>
              <TableCell>
                {token.scopes && token.scopes.length > 0 ? (
                  <div className="flex flex-wrap gap-1">
                    {token.scopes.map((scope) => (
                      <Badge key={scope} variant="secondary" className="font-mono text-xs">
                        {scope}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <span className="text-xs text-muted-foreground">Full access</span>
                )}
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {formatExpiry(token.expiresAt)}
              </TableCell>
              {canWrite && (
                <TableCell className="text-right">
                  <DeleteTokenDialog
                    tokenId={token.id}
                    identifier={token.identifier}
                  />
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
