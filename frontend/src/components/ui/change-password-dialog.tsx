import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "./button";
import { Input } from "./input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogTrigger,
} from "./dialog";
import { Label } from "./label";
import { updatePassword } from "@/lib/api";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { useToast } from "@/hooks/use-toast";
import { AxiosError } from "axios";
import { useState } from "react";

const PasswordFormSchema = z
  .object({
    current_password: z.string().min(1, "Current password is required"),
    new_password: z
      .string()
      .min(8, { message: "Password must be at least 8 characters long." })
      .max(64, "Password can not be longer than 64 characters"),
    confirm_password: z
      .string()
      .min(8, { message: "Password must be at least 8 characters long." })
      .max(64, "Password can not be longer than 64 characters"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Passwords don't match",
    path: ["confirm_password"],
  });

export function ChangePasswordDialog() {
  const form = useForm<z.infer<typeof PasswordFormSchema>>({
    resolver: zodResolver(PasswordFormSchema),
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  const { toast } = useToast();
  const [open, setOpen] = useState(false);

  const onSubmit = async (data: z.infer<typeof PasswordFormSchema>) => {
    const id = toast({
      title: "Update Password",
      description: "updating...",
    });
    try {
      await updatePassword(data.current_password, data.new_password);
      id.update({
        id: id.id,
        description: "Success!",
      });
      setOpen(false);
    } catch (e) {
      const error = e as AxiosError;
      const data = error?.response?.data as { error?: string };
      if (data && typeof data == "object" && "error" in data) {
        id.update({
          id: id.id,
          title: "Failed",
          description: data.error,
          variant: "destructive",
        });
      } else {
        id.update({
          id: id.id,
          title: "Failed",
          description: "Unkown error: " + e,
          variant: "destructive",
        });
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={(x) => setOpen(x)}>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full">
          Change Password
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Change Password</DialogTitle>
          <DialogDescription>
            Enter your current password and choose a new one.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="current_password"
              render={({ field }) => (
                <FormItem className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor="current_password" className="text-right">
                    Current
                  </Label>
                  <div className="col-span-3">
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </div>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="new_password"
              render={({ field }) => (
                <FormItem className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor="new_password" className="text-right">
                    New
                  </Label>
                  <div className="col-span-3">
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </div>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="confirm_password"
              render={({ field }) => (
                <FormItem className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor="confirm_password" className="text-right">
                    Confirm
                  </Label>
                  <div className="col-span-3">
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </div>
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type="submit">Save Password</Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
