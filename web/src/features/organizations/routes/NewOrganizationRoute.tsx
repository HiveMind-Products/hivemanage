import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Wordmark } from "@/components/brand/logo";
import { useCreateOrganization } from "../api/useCreateOrganization";
import { useForm } from "react-hook-form";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  createOrganizationSchema,
  CreateOrganizationSchema,
} from "@/typings/organizations";

export function NewOrganizationRoute() {
  const { mutate, isPending } = useCreateOrganization();

  const formMethods = useForm<CreateOrganizationSchema>({
    defaultValues: {
      name: "",
    },
    resolver: zodResolver(createOrganizationSchema),
  });

  function handleCreateOrganization(data: CreateOrganizationSchema) {
    mutate(data);
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-background p-6">
      <Wordmark markClassName="size-7" className="text-base" />
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Create your organization</CardTitle>
          <CardDescription>
            Organizations hold your files, logs, and team. You can rename it
            later.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...formMethods}>
            <form
              className="flex flex-col gap-6"
              onSubmit={formMethods.handleSubmit(handleCreateOrganization)}
            >
              <FormField
                control={formMethods.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Organization name</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder="Acme Roleplay"
                        autoFocus
                        disabled={isPending}
                      />
                    </FormControl>
                    <FormDescription>
                      Usually your community or company name.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" className="w-full" disabled={isPending}>
                {isPending ? "Creating…" : "Create organization"}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
