"use client";

import { useState } from "react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Chat, updateGroupChat, deleteChat as apiDeleteChat } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";
import { DialogProps, DialogTrigger } from "@radix-ui/react-dialog";
import { useAuth } from "@/provider/auth-provider";
import { Label } from "./label";
import { ConfirmDialog } from "./confirm";
import { AxiosError } from "axios";
import { useChat } from "@/provider/chat-provider";

export interface UpdateGroupProps extends DialogProps {
  chat: Chat;
}

export function UpdateGroup({
  chat,
  open,
  onOpenChange: onChange,
  ...props
}: UpdateGroupProps) {
  const [name, setName] = useState(chat.name);
  const { users } = useAuth();
  const members = chat.members.map((x) => users.find((u) => u.id == x)!);
  const { toast } = useToast();
  const { refresh } = useChat();
  const [close, setClose] = useState(false);

  const deleteChat = async () => {
    const id = toast({
      title: "Group Chat",
      description: "updating...",
    });
    try {
      await apiDeleteChat(chat.id!);
      id.update({
        id: id.id,
        description: "Success!",
      });
      refresh();
      setClose(true);
      onChange?.(false);
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

  const updateChat = async () => {
    const id = toast({
      title: "Group Chat",
      description: "updating...",
    });
    try {
      await updateGroupChat(chat.id!, name, chat.members);
      id.update({
        id: id.id,
        description: "Success!",
      });
      refresh();
    } catch (e) {
      const error = e as AxiosError;
      const data = error?.response?.data as { error?: string };
      if (data && "error" in data) {
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
      <Dialog {...props} open={close ? false : open} onOpenChange={onChange}>
        <DialogTrigger>Update Group</DialogTrigger>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Update Group Chat</DialogTitle>
            <p className="text-muted-foreground text-sm">
              {members.map((x) => `${x.first_name} ${x.last_name}`).join(",")}
            </p>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="flex justify-between items-center gap-2">
              <Label>Name: </Label>
              <Input
                placeholder="Group Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>

            <div className="flex justify-between items-center gap-2">
              <Label>Options: </Label>
              <ConfirmDialog
                title="Delete Group"
                description="This operation is not reversible"
                onConfirm={deleteChat}
              >
                <Button variant={"destructive"}>Delete Group</Button>
              </ConfirmDialog>
            </div>
          </div>
          <DialogClose>
            <Button className="w-full" onClick={updateChat}>
              Update
            </Button>
          </DialogClose>
        </DialogContent>
      </Dialog>
    </>
  );
}
