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
}

export function DeleteTokenDialog(props: DeleteTokenDialogProps) {
  const { tokenId } = props;

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
          className="h-8 w-8 text-muted-foreground hover:text-destructive transition-colors"
        >
          <Trash size={16} />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Token</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete this token? This action cannot be
            undone and any applications using this token will lose access.
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
            {isPending ? "Deleting..." : "Delete Token"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
