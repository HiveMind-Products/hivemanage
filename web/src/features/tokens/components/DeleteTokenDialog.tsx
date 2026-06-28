import { Button } from "@/components/ui/button";
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
import { Trash } from "lucide-react";
import { useDeleteToken } from "../api/useDeleteToken";
import { useQueryClient } from "@tanstack/react-query";
import { QueryKeys } from "@/typings/query";
import { useTransition } from "react";
import { useParams } from "react-router";
import { Params } from "@/typings/router";

interface DeleteTokenDialogProps {
  tokenId: number;
  identifier?: string;
}

export function DeleteTokenDialog(props: DeleteTokenDialogProps) {
  const { tokenId, identifier } = props;

  const [isPending, startTransition] = useTransition();

  const params = useParams<Params>();
  const queryClient = useQueryClient();
  const { mutateAsync } = useDeleteToken();

  async function handleDelete() {
    startTransition(async () => {
      await mutateAsync(tokenId);

      startTransition(() => {
        queryClient.invalidateQueries({ queryKey: [QueryKeys.Tokens, params.organizationId] });
      });
    });
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          size="icon"
          variant="ghost"
          aria-label="Delete token"
          className="size-8 text-muted-foreground transition-colors hover:text-destructive"
        >
          <Trash className="size-4" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            Delete {identifier ? `"${identifier}"` : "token"}?
          </DialogTitle>
          <DialogDescription>
            This permanently revokes the token. Any application using it will
            immediately lose access. This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={isPending}
          >
            {isPending ? "Deleting…" : "Delete token"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
