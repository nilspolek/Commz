import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { useAuth } from "@/provider/auth-provider";
import { useNavigate } from "react-router";

const FormSchema = z
  .object({
    email: z.string().email({
      message: "Must be a valid E-Mail address.",
    }),
    password: z
      .string()
      .min(8, { message: "Password must be at least 8 characters long." })
      .max(64, "Password can not be longer than 64 characters"),
    repeat: z
      .string()
      .min(8, { message: "Password must be at least 8 characters long." })
      .max(64, "Password can not be longer than 64 characters"),
    first_name: z.string().min(1),
    last_name: z.string().min(1),
  })
  .refine((data) => data.password === data.repeat, {
    message: "Passwords don't match",
    path: ["repeat"],
  });

function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();

  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      email: "",
    },
  });

  async function onSubmit(data: z.infer<typeof FormSchema>) {
    const success = await register(data);
    if (success) navigate("/");
  }

  return (
    <>
      <div className="w-dvw h-dvh flex items-center justify-center flex-col gap-4">
        <h1 className="font-bold text-3xl">Register a new account</h1>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="max-w-[450px] w-full"
          >
            <Card className="p-4 space-y-2">
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
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Password"
                        {...field}
                        type="password"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="repeat"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password Again</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Password"
                        {...field}
                        type="password"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className="text-right font-bold text-primary">
                <a href="/login">Already have an account?</a>
              </div>
              <Button type="submit" className="w-full">
                Register
              </Button>
            </Card>
          </form>
        </Form>
      </div>
    </>
  );
}

export default Register;
