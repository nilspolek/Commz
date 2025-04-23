import { useAuth } from "@/provider/auth-provider";
import { useChat } from "@/provider/chat-provider";
import { Input } from "./input";
import { Button } from "./button";
import { EllipsisVertical, Loader2, Search } from "lucide-react";
import {
  createDirectChat,
  getMessages,
  Message,
  Summary,
  updateGroupChat,
} from "@/lib/api";
import { useEffect, useState } from "react";
import { cn, getPictureUrl } from "@/lib/utils";
import DotPattern from "./dot-pattern";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import { ConfirmDialog } from "./confirm";
import { UpdateGroup } from "./update-group";
import { useDialog } from "@/provider/dialog-provider";
import { ManageGroupMember } from "./manage-group-members";
import { useToast } from "@/hooks/use-toast";
import { AxiosError } from "axios";
import { ChatBubble, SummaryBubble } from "./chat-bubble";
import { ChatInput } from "./chat-input";

export const Chat = () => {
  const { chat, setChat, sendMessage, refresh } = useChat();
  const { toast } = useToast();
  const { user, users } = useAuth();
  const [search, setSearch] = useState<string>();
  const { create } = useDialog();
  const [originalMessages, setMessages] = useState<Message[]>([]);
  const messages = originalMessages.filter((x) =>
    x.content.includes(search || "")
  );
  const [summary, setSummary] = useState<Summary | undefined>();
  const [reply, setReply] = useState<Message>();
  const [canLoadMessages, setCanLoadMessages] = useState(true);
  const [loadingMessages, setLoadingMessages] = useState(false);

  const loadMessages = async () => {
    if (!chat) return;
    try {
      setLoadingMessages(true);
      const newMessages = await getMessages(chat.id!, 30, messages.length);
      if (newMessages.length < 30) {
        setCanLoadMessages(false);
      }
      setMessages([...newMessages, ...messages]);
    } finally {
      setLoadingMessages(false);
    }
  };

  useEffect(() => {
    if (summary && summary.chat_id != chat?.id) {
      setSummary(undefined);
    }
  }, [chat, summary]);

  useEffect(() => {
    setCanLoadMessages(true);
    if (!chat || !chat.messages) {
      setMessages([]);
      return;
    }
    setReply(undefined);
    setMessages(chat.messages);
  }, [chat]);

  if (!user || !chat)
    return (
      <div className="flex flex-col w-full h-full">
        <div className="border-border border-b h-20 bg-card w-full"></div>
        <div className="relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg bg-background my-auto">
          <div className="z-10 whitespace-pre-wrap text-center text-5xl font-medium tracking-tighter text-black dark:text-white">
            No active chat yet.
            <p className="text-lg tracking-normal opacity-50">
              Start a new chat, by writing one of the persons on the left.
            </p>
          </div>
          <DotPattern
            className={cn(
              "[mask-image:radial-gradient(500px_circle_at_center,white,transparent)]"
            )}
          />
        </div>
      </div>
    );
  const otherUserId = chat.members.filter((x) => x != user.id).at(0);
  const otherUser = users.find((x) => x.id == otherUserId);

  const submit = async (
    message: string,
    command?: string,
    media?: string[]
  ) => {
    if (message.length == 0 && (!media || media.length == 0)) return;
    try {
      if (!chat.id && otherUserId) {
        // chat does not exists yet, we have to create it first
        const result = await createDirectChat(otherUserId, "");
        setChat(result);
        sendMessage(result.id!, message, command, media, reply?.id);
        return;
      }
      const result = await sendMessage(
        chat.id!,
        message,
        command,
        media,
        reply?.id
      );

      setMessages([...messages, result]);
    } catch (e) {
      const error = e as AxiosError;
      const data = error?.response?.data as { error?: string };
      if (data && typeof data == "object" && "error" in data) {
        toast({
          title: "Failed",
          description: data.error,
          variant: "destructive",
        });
      } else {
        toast({
          title: "Failed",
          description: "Unkown error: " + e,
          variant: "destructive",
        });
      }
    }
  };

  const leaveGroup = async () => {
    if (!user) return;
    const id = toast({
      title: "Group Chat",
      description: "leaving...",
    });
    try {
      await updateGroupChat(
        chat.id!,
        chat.name,
        chat.members.filter((u) => u != user.id)
      );
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

  const reverseMessages = [...messages].reverse();

  return (
    <div className="w-full h-dvh flex flex-col">
      <div className="border-b border-border h-20 bg-card inline-flex w-full items-center gap-2 p-2 shrink-0 relative">
        <div
          className={cn(
            "absolute right-0 z-40 border border-border rounded-md bg-card p-2 top-full transition-all duration-500 w-48 ml-auto",
            search !== undefined
              ? "opacity-100 scale-100 -translate-y-0 pointer-events-auto"
              : "scale-100 -translate-y-10 opacity-0 pointer-events-none"
          )}
        >
          <div className="relative">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              className="pl-8"
              placeholder="Search chat..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onBlur={() => setSearch(undefined)}
            />
          </div>
        </div>
        {((chat.members.length == 2 &&
          chat.name == "Direct Chat" &&
          otherUser) ||
          (!chat.id && otherUser)) && (
          <>
            <img
              src={getPictureUrl(otherUser)}
              alt={otherUser.first_name}
              className="w-10 h-10 rounded-full mr-4"
            />
            <div className="font-bold text-2xl">
              {otherUser.first_name} {otherUser.last_name}
            </div>
          </>
        )}
        {chat.name != "Direct Chat" && chat.name != "AI" && chat.id && (
          <div>
            <h2 className="font-semibold text-lg">{chat.name}</h2>
            <p className="text-sm text-muted-foreground font-semibold">
              {chat.members
                .map((x) => users.find((u) => u.id == x))
                .filter(Boolean)
                .map((u) => `${u?.first_name} ${u?.last_name}`)
                .join(", ")}
            </p>
          </div>
        )}
        {chat.name == "AI" && (
          <div>
            <h2 className="font-semibold text-lg">{chat.name}</h2>
            <p className="text-sm text-muted-foreground font-semibold">
              Chat with the AI
            </p>
          </div>
        )}
        <div className="ml-auto px-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost">
                <EllipsisVertical></EllipsisVertical>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56">
              {chat.name != "Direct Chat" && chat.id && chat.name != "AI" && (
                <>
                  <DropdownMenuLabel>Settings</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem
                      onClick={() =>
                        create((props) => (
                          <UpdateGroup {...props} chat={chat} />
                        ))
                      }
                    >
                      Update Group
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() =>
                        create((props) => (
                          <ManageGroupMember {...props} chat={chat} />
                        ))
                      }
                    >
                      Add Member
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onSelect={(e) => e.preventDefault()}
                      className="text-destructive"
                    >
                      <ConfirmDialog onConfirm={leaveGroup} title="Confrim">
                        Leave Group
                      </ConfirmDialog>
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                </>
              )}
              <DropdownMenuGroup>
                <DropdownMenuLabel>Options</DropdownMenuLabel>
                <DropdownMenuSub>
                  <DropdownMenuItem onClick={() => setSearch("")}>
                    <Search /> Search
                  </DropdownMenuItem>
                </DropdownMenuSub>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
      <div className="flex flex-col-reverse gap-2 p-2 h-full overflow-y-auto relative overflow-x-hidden">
        <DotPattern
          className={cn(
            "[mask-image:radial-gradient(500px_circle_at_center,white,transparent)]"
          )}
        />

        {summary && summary.chat_id == chat.id && (
          <SummaryBubble summary={summary} />
        )}

        {chat.name == "AI" && chat.messages?.at(-1)?.sender == user.id && (
          <div className="inline-flex gap-2 max-w-[66%] z-10">
            <img
              src={"https://ui-avatars.com/api/?name=AI"}
              alt={"ai"}
              className="w-10 h-10 rounded-full"
            />
            <div
              className={cn(
                "bg-card bg-opacity-10 p-2 min-w-28 rounded-md relative pr-12",
                "rounded-tl-none"
              )}
            >
              Generating answer
              <div className="inline-flex gap-1 ml-2">
                <span className="bg-muted-foreground w-1 h-1 rounded-full animate-pulse delay-0 duration-1000" />
                <span className="bg-muted-foreground w-1 h-1 rounded-full animate-pulse delay-100 duration-1000" />
                <span className="bg-muted-foreground w-1 h-1 rounded-full animate-pulse delay-200 duration-1000" />
                <span className="bg-muted-foreground w-1 h-1 rounded-full animate-pulse delay-300 duration-1000" />
              </div>
            </div>
          </div>
        )}
        {!chat.messages && (
          <div className="bg-muted mx-auto p-1 rounded-md text-muted-foreground">
            Start a new chat by sending a message.
          </div>
        )}
        {reverseMessages.map((_, i) => (
          <ChatBubble
            index={i}
            messages={reverseMessages}
            key={i}
            users={users}
            user={user}
            setReply={setReply}
          />
        ))}
        {canLoadMessages && messages.length > 0 && messages.length >= 30 && (
          <div className="w-full flex justify-center">
            <Button variant="ghost" onClick={loadMessages}>
              Load more messages
              {loadingMessages && <Loader2 className="animate-spin" />}
            </Button>
          </div>
        )}
      </div>
      <ChatInput
        submit={submit}
        setSummary={setSummary}
        reply={reply}
        setReply={setReply}
      />
    </div>
  );
};
