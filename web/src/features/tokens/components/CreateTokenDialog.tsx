import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { type TokenParams, tokenSchema } from "@/typings/token";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useCreateToken } from "../api/useCreateToken";
import { useState, useTransition } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AlertTriangle, Check, Copy, KeyRound } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { QueryKeys } from "@/typings/query";
import { useParams } from "react-router";
import { Params } from "@/typings/router";
import { useCopyToClipboard } from "@/hooks/use-copy";

export function CreateTokenDialog() {
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const { copied, copy } = useCopyToClipboard();

  const { mutateAsync, reset, data, isSuccess } = useCreateToken();
  const form = useForm<TokenParams>({
    defaultValues: {
      type: "media",
      identifier: "",
    },
    resolver: zodResolver(tokenSchema),
  });

  const params = useParams<Params>();
  const queryClient = useQueryClient();

  function handleOnSubmit(data: TokenParams) {
    startTransition(async () => {
      await mutateAsync(data);

      startTransition(() => {
        queryClient.invalidateQueries({ queryKey: [QueryKeys.Tokens, params.organizationId] });
      });
    });
  }

  function resetForm() {
    form.reset();
    reset();
  }

  function handleInteractOutside(e: Event) {
    e.preventDefault();
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <KeyRound className="size-4" />
          Create token
        </Button>
      </DialogTrigger>
      <DialogContent
        id="create-token-dialog"
        onInteractOutside={handleInteractOutside}
        onPointerDownOutside={handleInteractOutside}
        onAnimationEnd={resetForm}
        className="sm:max-w-[440px]"
      >
        <DialogHeader>
          <DialogTitle>{isSuccess ? "Token created" : "Create token"}</DialogTitle>
          <DialogDescription>
            {isSuccess
              ? "Copy your token now — for your security, it won't be shown again."
              : "Create a new token to authenticate with the platform's API."}
          </DialogDescription>
        </DialogHeader>
        {isSuccess && data ? (
          <div className="space-y-4 pt-2">
            <div className="flex items-stretch gap-2">
              <code className="min-w-0 flex-1 break-all rounded-md border bg-muted px-3 py-2.5 font-mono text-sm">
                {data.token}
              </code>
              <Button
                type="button"
                variant="outline"
                size="icon"
                aria-label="Copy token"
                className="h-auto shrink-0"
                onClick={() => copy(data.token, "Token copied to clipboard")}
              >
                {copied ? (
                  <Check className="size-4 text-success" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>
            <div className="flex items-start gap-2.5 rounded-md border border-warning/30 bg-warning/10 p-3 text-xs text-foreground">
              <AlertTriangle className="mt-0.5 size-4 shrink-0 text-warning" />
              <span>
                This is the only time the full token is displayed. Store it
                somewhere safe before closing this dialog.
              </span>
            </div>
            <DialogFooter>
              <Button onClick={() => setOpen(false)} className="w-full">
                I've copied it
              </Button>
            </DialogFooter>
          </div>
        ) : (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(handleOnSubmit)}
              className="space-y-4 pt-2"
            >
              <FormField
                control={form.control}
                name="identifier"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Identifier</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder="e.g. Production website"
                        type="text"
                        autoComplete="off"
                        autoFocus
                      />
                    </FormControl>
                    <FormDescription>
                      A name to help you recognize this token later.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <DialogFooter className="pt-2">
                <Button type="submit" disabled={isPending} className="w-full">
                  {isPending ? "Creating…" : "Create token"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  );
}
