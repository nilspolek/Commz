import { useToast } from "@/hooks/use-toast";
import { LoginRequest, RegisterRequest, User } from "@/lib/api";
import * as api from "@/lib/api";
import { AxiosError } from "axios";
import React, { useContext, useEffect, useState } from "react";

interface AuthContextType {
  user?: User;
  users: User[];
  login: (data: LoginRequest) => Promise<boolean>;
  register: (data: RegisterRequest) => Promise<boolean>;
  logout: () => Promise<boolean>;
  loggedIn: boolean;
  loading: boolean;
  refresh: () => void;
}

const AuthContext = React.createContext<AuthContextType | null>(null);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User>();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    const load = async () => {
      try {
        setUsers(await api.getUsers());
        setUser(await api.getUser());
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const login = async (data: LoginRequest) => {
    const id = toast({
      title: "Login",
      description: "loading...",
    });
    try {
      const result = await api.login(data);
      setUser(result);
      id.update({
        id: id.id,
        description: "Success!",
      });
    } catch (e) {
      const error = e as AxiosError;
      const data = error?.response?.data as { error?: string };
      id.update({
        id: id.id,
        title: "Failed",
        description: data.error,
        variant: "destructive",
      });
      return false;
    }
    return true;
  };

  const logout = async () => {
    await api.userLogout();
    setUser(undefined);
    return true;
  };

  const register = async (data: RegisterRequest) => {
    const id = toast({
      title: "Register",
      description: "loading...",
    });
    try {
      const result = await api.register(data);
      setUser(result);
      id.update({
        id: id.id,
        description: "Success!",
      });
    } catch (e) {
      const error = e as AxiosError;
      const data = error?.response?.data as { error?: string };
      id.update({
        id: id.id,
        title: "Failed",
        description: data.error,
        variant: "destructive",
      });
    }
    return true;
  };

  return (
    <AuthContext.Provider
      value={{
        refresh: () => {
          api.getUser().then((x) => setUser(x));
        },
        loading,
        user,
        users,
        login,
        logout,
        register,
        loggedIn: user != null || user !== undefined,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error("can only useAuth in AuthProvider");
  return context;
};
