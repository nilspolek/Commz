import { Loader2, Settings } from "lucide-react";
import { Button } from "./button";
import { Input } from "./input";
import {
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
  Sheet,
} from "./sheet";
import { Separator } from "./separator";
import {
  updateProfile,
  User,
  uploadImage as apiUploadImage,
  getVersion,
} from "@/lib/api";
import { ConfirmDialog } from "./confirm";
import { ChangePasswordDialog } from "./change-password-dialog";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "./form";
import { useToast } from "@/hooks/use-toast";
import { AxiosError } from "axios";
import { useAuth } from "@/provider/auth-provider";
import { useEffect, useState } from "react";
import { cn, getPictureUrl } from "@/lib/utils";

const FormSchema = z.object({
  email: z.string().email({
    message: "Must be a valid E-Mail address.",
  }),
  first_name: z.string().min(1),
  last_name: z.string().min(1),
});

const UpdateProfile = ({ user }: { user: User }) => {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      email: user.email,
      first_name: user.first_name,
      last_name: user.last_name,
    },
  });

  const { toast } = useToast();
  const { refresh, logout } = useAuth();
  const [image, setImage] = useState(user.picture);
  const [uploading, setUploading] = useState(false);

  const [versions, setVersions] = useState<{ name: string; version: string }[]>(
    []
  );

  useEffect(() => {
    const loadVersions = async () => {
      const services = ["auth", "ai", "chat", "media", "gateway"];
      const result = await Promise.all(services.map((x) => getVersion(x)));
      setVersions([
        ...services.map((x, i) => ({
          name: x + "-service",
          version: result[i],
        })),
        { name: "frontend", version: "1.2.0" },
      ]);
    };
    loadVersions();
  }, []);

  async function uploadImage() {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = "image/*";
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;

      const reader = new FileReader();
      reader.readAsArrayBuffer(file);
      reader.onload = async () => {
        try {
          setUploading(true);
          const bytes = new Uint8Array(reader.result as ArrayBuffer);
          // You can now use the bytes array for further processing
          const result = await apiUploadImage(bytes, file.type);
          setImage(result);
        } catch {
          toast({
            title: "Upload failed",
            description: "Make sure to select an image",
            variant: "destructive",
          });
        } finally {
          setUploading(false);
        }
      };
    };
    input.click();
  }

  async function update(data: z.infer<typeof FormSchema>) {
    const id = toast({
      title: "Update Profile",
      description: "updating...",
    });
    try {
      await updateProfile({ ...data, id: user.id, picture: image });
      id.update({
        id: id.id,
        description: "Success!",
      });
      refresh();
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
  }

  return (
    <Sheet>
      <SheetTrigger>
        <Button variant="ghost">
          <Settings />
        </Button>
      </SheetTrigger>
      <SheetContent className="h-full flex flex-col gap-1">
        <SheetHeader>
          <SheetTitle>Profile</SheetTitle>
          <SheetDescription>
            Make changes to your profile here. Click save when you're done.
          </SheetDescription>
        </SheetHeader>

        <div className="w-full flex flex-col items-center gap-y-4 my-4 relative">
          <img
            src={getPictureUrl(image || user)}
            alt={user.first_name}
            onClick={uploadImage}
            className={cn(
              "w-16 h-16 rounded-full mr-2 cursor-pointer transition-all",
              uploading && "opacity-30"
            )}
          />
          <Loader2
            className={cn(
              "absolute top-5 text-primary animate-spin",
              !uploading && "hidden"
            )}
          />
          {user.first_name} {user.last_name}
        </div>

        <Form {...form}>
          <form onSubmit={(e) => e.preventDefault()}>
            <div className="grid gap-2">
              <div className="inline-flex gap-2 w-full">
                <FormField
                  control={form.control}
                  name="first_name"
                  render={({ field }) => (
                    <FormItem className="w-full">
                      <FormLabel>First Name</FormLabel>
                      <FormControl>
                        <Input placeholder="First Name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="last_name"
                  render={({ field }) => (
                    <FormItem className="w-full">
                      <FormLabel>Last Name</FormLabel>
                      <FormControl>
                        <Input placeholder="Last Name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>E-Mail</FormLabel>
                    <FormControl>
                      <Input placeholder="your@email.com" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Separator />
              <div className="flex justify-end">
                <ChangePasswordDialog />
              </div>
              <SheetFooter className="w-full">
                <Button className="w-full" variant="ghost" onClick={logout}>
                  Log out
                </Button>
                <ConfirmDialog
                  title="Confirm Changes"
                  description="Please confirm your profile changes"
                  confirmText="Save"
                  onConfirm={async () => {
                    const isValid = await form.trigger();
                    if (isValid) {
                      await update(form.getValues());
                    }
                  }}
                >
                  <Button className="w-full">Save changes</Button>
                </ConfirmDialog>
              </SheetFooter>
            </div>
          </form>
        </Form>
        <div className="mt-auto">
          <table className="w-full text-muted-foreground text-sm">
            <tbody>
              {versions.map((x, i) => (
                <tr key={i}>
                  <td className="w-32">{x.name}</td>
                  <td>v{x.version}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </SheetContent>
    </Sheet>
  );
};

export default UpdateProfile;
