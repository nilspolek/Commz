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
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/provider/auth-provider";
import { useNavigate } from "react-router";
import { useEffect } from "react";

const FormSchema = z.object({
  email: z.string().email({
    message: "Must be a valid E-Mail address.",
  }),
  password: z
    .string()
    .min(8, { message: "Password must be at least 8 characters long." })
    .max(64, "Password can not be longer than 64 characters"),
});

function Login() {
  const { login, loggedIn } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (loggedIn) navigate("/");
  }, [loggedIn, navigate]);

  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      email: "",
    },
  });

  async function onSubmit(data: z.infer<typeof FormSchema>) {
    const success = await login(data);
    if (success) navigate("/");
  }

  return (
    <>
      <div className="w-dvw h-dvh flex items-center justify-center flex-col gap-4">
        <h1 className="font-bold text-3xl">Sign into your account</h1>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="max-w-[450px] w-full"
          >
            <Card className="p-4 space-y-4">
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
              <div className="text-right font-bold inline-flex justify-between w-full items-center">
                <div className="inline-flex gap-1 items-center">
                  <Checkbox />
                  <Label>Remember me</Label>
                </div>
                <a href="/register" className="text-primary">
                  Don't have an account yet?
                </a>
              </div>
              <Button type="submit" className="w-full">
                Login
              </Button>
            </Card>
          </form>
        </Form>
      </div>
    </>
  );
}

export default Login;
