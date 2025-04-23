"use client";

import { useState } from "react";
import { Search, Plus } from "lucide-react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useAuth } from "@/provider/auth-provider";
import { Chat, updateGroupChat } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";
import { AxiosError } from "axios";
import { DialogProps } from "@radix-ui/react-dialog";
import { useChat } from "@/provider/chat-provider";
import { getPictureUrl } from "@/lib/utils";

export interface ManageGroupMemberProps extends DialogProps {
  chat: Chat;
}

export function ManageGroupMember({ chat, ...props }: ManageGroupMemberProps) {
  const [selectedUsers, setSelectedUsers] = useState<string[]>(chat.members);
  const { users: originalUsers, user } = useAuth();
  const { refresh } = useChat();
  const [search, setSearch] = useState("");

  const users = [...originalUsers]
    .filter(
      (x) =>
        x.first_name.toLowerCase().includes(search.toLowerCase()) ||
        x.last_name.toLowerCase().includes(search.toLowerCase())
    )
    .filter((x) => x.id != user?.id);
  const { toast } = useToast();

  const updateGroup = async () => {
    if (!user) return;
    const id = toast({
      title: "Group Chat",
      description: "updating...",
    });
    try {
      await updateGroupChat(chat.id!, chat.name, selectedUsers);
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
  };

  return (
    <>
      <Dialog {...props}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Manage Group Members</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="relative">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search users"
                className="pl-8"
                onChange={(e) => setSearch(e.target.value)}
                value={search}
              />
            </div>
            <ScrollArea className="h-[200px] rounded-md border p-2">
              {users.map((user, i) => (
                <div
                  key={i}
                  className="flex items-center gap-2 rounded-lg p-2 hover:bg-muted"
                >
                  <img
                    src={getPictureUrl(user)}
                    alt={user.first_name}
                    className="w-8 h-8 rounded-full mr-2"
                  />
                  <div className="flex-1">
                    {user.first_name} {user.last_name}
                  </div>
                  <Button
                    size="icon"
                    variant={
                      selectedUsers.includes(user.id) ? "default" : "outline"
                    }
                    onClick={() => {
                      setSelectedUsers((prev) =>
                        prev.find((i) => i == user.id)
                          ? prev.filter((id) => id !== user.id)
                          : [...prev, user.id]
                      );
                    }}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
              ))}

              {users.length == 0 && <>No users found.</>}
            </ScrollArea>
            <DialogClose>
              <Button className="w-full" onClick={updateGroup}>
                Update Group
              </Button>
            </DialogClose>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
