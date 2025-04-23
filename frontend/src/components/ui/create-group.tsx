"use client";

import { useState } from "react";
import { Users, Search, Plus } from "lucide-react";
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
import { createGroupChat, User } from "@/lib/api";
import { InteractiveHoverButton } from "./interactive-hover-button";
import { useToast } from "@/hooks/use-toast";
import { AxiosError } from "axios";
import { useChat } from "@/provider/chat-provider";
import { getPictureUrl } from "@/lib/utils";

export function CreateGroupChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedUsers, setSelectedUsers] = useState<User[]>([]);
  const { users: originalUsers, user } = useAuth();
  const [search, setSearch] = useState("");
  const [name, setName] = useState("");
  const { refresh } = useChat();

  const users = [...originalUsers]
    .filter(
      (x) =>
        x.first_name.toLowerCase().includes(search.toLowerCase()) ||
        x.last_name.toLowerCase().includes(search.toLowerCase())
    )
    .filter((x) => x.id != user?.id);
  const { toast } = useToast();

  const createChat = async () => {
    if (!user) return;
    const id = toast({
      title: "Group Chat",
      description: "creating...",
    });
    try {
      await createGroupChat(name, [...selectedUsers.map((x) => x.id), user.id]);
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
      <InteractiveHoverButton
        onClick={() => setIsOpen(true)}
        first={
          <div className="flex gap-1 flex-nowrap items-center relative ml-2 text-sm">
            <Users className="h-4 w-4 absolute -left-5" />
            <span className="max-w-full overflow-hidden text-ellipsis text-nowrap">
              Create Group Chat
            </span>
          </div>
        }
        second="Click here"
        className="mx-auto"
      />
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Create Group Chat</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <Input
              placeholder="Group Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
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
                      selectedUsers.includes(user) ? "default" : "outline"
                    }
                    onClick={() => {
                      setSelectedUsers((prev) =>
                        prev.find((i) => i.id == user.id)
                          ? prev.filter((id) => id.id !== user.id)
                          : [...prev, user]
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
              <Button onClick={createChat}>Create Group</Button>
            </DialogClose>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
