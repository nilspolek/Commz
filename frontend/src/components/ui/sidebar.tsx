import { useChat } from "@/provider/chat-provider";
import { Input } from "./input";
import { useAuth } from "@/provider/auth-provider";
import { useState } from "react";
import { Chat, Message, User } from "@/lib/api";
import { cn, getChatMember, getPictureUrl } from "@/lib/utils";
import { Skeleton } from "./skeleton";
import { format } from "date-fns";
import { Search } from "lucide-react";
import { Button } from "./button";
import { CreateGroupChat } from "./create-group";
import UpdateProfile from "./update-profile";

export const Sidebar = () => {
  const {
    chats: originalChats,
    setChat,
    loading: chatLoading,
    chat,
  } = useChat();
  const { users: originalUsers, user, loading: authLoading } = useAuth();

  const loading = authLoading || chatLoading;

  const [search, setSearch] = useState("");

  const users = originalUsers.filter(
    (x) =>
      x.first_name.toLowerCase().includes(search.toLowerCase()) ||
      x.last_name.toLowerCase().includes(search.toLowerCase())
  );
  const chats = originalChats.filter(
    (x) =>
      x.name.toLowerCase().includes(search.toLowerCase()) ||
      x.name.toLowerCase().includes(search.toLowerCase())
  );

  const directChats = chats.filter(
    (x) => x.name == "Direct Chat" || x.name == "AI"
  );
  const groupChats = chats.filter(
    (x) => x.name != "Direct Chat" && x.name != "AI"
  );

  if (loading)
    return (
      <>
        <div className="h-full w-1/4 border-border border-r">
          <div className="w-full h-20 border-b border-border inline-flex items-center font-medium p-2 text-2xl bg-card">
            Chats
          </div>
          <div className="p-4 border-border border-b">
            <Input
              placeholder="Search for a new chat"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          {[...Array(10)].map((_, i) => (
            <div
              key={i}
              className="w-full h-16 inline-flex p-2 items-center border-border border-b hover:bg-border transition-all cursor-pointer overflow-hidden"
            >
              <Skeleton className="rounded-full mr-4 w-10 h-10 shrink-0" />
              <div className="overflow-hidden text-nowrap">
                <div className="overflow-hidden text-ellipsis">
                  <Skeleton className="w-32 h-4" />
                </div>
                <div className="text-sm opacity-50 overflow-hidden text-ellipsis">
                  <Skeleton className="h-4 w-20" />
                </div>
              </div>
            </div>
          ))}
        </div>
      </>
    );
  if (!user) return <>You have to be logged in </>;

  const hasChat = (user_id: string) =>
    chats.filter((x) => x.members.includes(user_id)).length >= 1;

  const countNotReadMessage = (messages?: Message[]) => {
    return (
      messages?.filter((x) => x.read === false && x.sender != user.id).length ||
      0
    );
  };

  const ChatDetails = ({
    user,
    messages,
    last_active,
    chat,
  }: {
    user?: User | string[];
    messages?: Message[];
    last_active?: Date;
    chat?: Chat;
  }) => {
    const missingMessages =
      countNotReadMessage(chat?.messages || messages) || 0;

    if (Array.isArray(user)) {
      const members = user.map((x) => originalUsers.find((u) => u.id == x)!);

      return (
        <>
          <img
            src={getPictureUrl(members[0])}
            alt={members[0].first_name}
            className="w-8 h-8 rounded-full mr-2"
          />
          <div className="overflow-hidden text-nowrap text-sm mr-auto">
            <div className="overflow-hidden text-ellipsis font-medium">
              {chat?.name ||
                members.map((x) => `${x.first_name} ${x.last_name}`).join(", ")}
            </div>
            <div className="text-sm opacity-50 overflow-hidden text-ellipsis">
              {messages ? messages.at(-1)?.content : "Start a new chat"}
            </div>
          </div>
          {last_active && missingMessages == 0 && (
            <div className="text-sm ml-auto opacity-50">
              {format(last_active, "HH:mm")}
            </div>
          )}
          {missingMessages > 0 && (
            <div className="bg-primary rounded-full shrink-0 w-4 h-4 flex items-center text-sm font-medium text-primary-foreground">
              <div className="mx-auto">{missingMessages}</div>
            </div>
          )}
        </>
      );
    }

    if (!user) {
      return (
        <>
          <img
            src={"https://ui-avatars.com/api/?name=AI"}
            alt={"AI"}
            className="w-8 h-8 rounded-full mr-2"
          />
          <div className="overflow-hidden text-nowrap text-sm mr-auto">
            <div className="overflow-hidden text-ellipsis font-medium">AI</div>
            <div className="text-sm opacity-50 overflow-hidden text-ellipsis">
              {messages ? messages.at(-1)?.content : "Start a new chat"}
            </div>
          </div>
          {last_active && missingMessages == 0 && (
            <div className="text-sm ml-auto opacity-50">
              {format(last_active, "HH:mm")}
            </div>
          )}
          {missingMessages > 0 && (
            <div className="bg-primary rounded-full shrink-0 w-4 h-4 flex items-center text-sm font-medium text-primary-foreground">
              <div className="mx-auto">{missingMessages}</div>
            </div>
          )}
        </>
      );
    }

    return (
      <>
        <img
          src={getPictureUrl(user)}
          alt={user.first_name}
          className="w-8 h-8 rounded-full mr-2"
        />
        <div className="overflow-hidden text-nowrap text-sm mr-auto">
          <div className="overflow-hidden text-ellipsis font-medium">
            {user.first_name} {user.last_name}
          </div>
          <div className="text-sm opacity-50 overflow-hidden text-ellipsis">
            {messages ? messages.at(-1)?.content : "Start a new chat"}
          </div>
        </div>
        {last_active && missingMessages == 0 && (
          <div className="text-sm ml-auto opacity-50">
            {format(last_active, "HH:mm")}
          </div>
        )}
        {missingMessages > 0 && (
          <div className="bg-primary rounded-full shrink-0 w-4 h-4 flex items-center text-sm font-medium text-primary-foreground">
            <div className="mx-auto">{missingMessages}</div>
          </div>
        )}
      </>
    );
  };

  return (
    <>
      <div className="h-full w-1/4 border-border border-r">
        <div className="w-full h-20 border-b border-border inline-flex items-center font-medium p-2 text-2xl bg-card justify-between">
          Chats
          <UpdateProfile user={user} />
        </div>
        <div className="p-4">
          <div className="relative px-4">
            <Search className="absolute left-6 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              className="pl-8"
              placeholder="Search chats..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>
        <div className="flex items-center">
          <CreateGroupChat />
        </div>
        <div className="flex flex-col px-4 space-y-1">
          <h2 className="mb-2 px-2 text-xs font-semibold text-muted-foreground pt-4">
            Direct Messages
          </h2>
          {directChats.map((x, i) => (
            <Button
              key={i}
              className={cn(
                "w-full h-10 inline-flex p-2 items-center hover:bg-border transition-all cursor-pointer overflow-hidde text-left",
                chat?.id == x.id && "bg-border"
              )}
              variant={"ghost"}
              onClick={() => {
                setChat(x);
              }}
            >
              <ChatDetails
                user={getChatMember(x, user, originalUsers)}
                messages={x.messages}
                last_active={x.last_active}
              />
            </Button>
          ))}
          {directChats.length == 0 && (
            <div className="text-xs font-medium text-muted-foreground px-2">
              No Direct Chats yet.
            </div>
          )}

          <h2 className="mb-2 px-2 text-xs font-semibold text-muted-foreground pt-4">
            Group Chats
          </h2>
          {groupChats.map((x, i) => (
            <Button
              key={i}
              className={cn(
                "w-full h-10 inline-flex p-2 items-center hover:bg-border transition-all cursor-pointer overflow-hidde text-left",
                chat?.id == x.id && "bg-border"
              )}
              variant={"ghost"}
              onClick={() => {
                setChat(x);
              }}
            >
              <ChatDetails
                user={x.members}
                messages={x.messages}
                last_active={x.last_active}
                chat={x}
              />
            </Button>
          ))}
          {groupChats.length == 0 && (
            <div className="text-xs font-medium text-muted-foreground px-2">
              No Group Chats yet.
            </div>
          )}

          {users.filter((x) => !hasChat(x.id)).filter((x) => x.id != user.id)
            .length > 0 && (
            <h2 className="mb-2 px-2 text-xs font-semibold text-muted-foreground pt-4">
              Users
            </h2>
          )}
          {users
            .filter((x) => !hasChat(x.id))
            .filter((x) => x.id != user.id)
            .map((x, i) => (
              <Button
                key={i}
                className={cn(
                  "w-full h-10 inline-flex p-2 items-center hover:bg-border transition-all cursor-pointer overflow-hidde text-left",
                  chat?.members.includes(x.id) && "bg-border"
                )}
                variant={"ghost"}
                onClick={() => {
                  const chat: Chat = {
                    members: [user.id, x.id],
                    name: "",
                    createdAt: new Date(),
                    last_active: new Date(),
                  };
                  setChat(chat);
                }}
              >
                <ChatDetails user={x} />
              </Button>
            ))}
        </div>
        {users.length == 0 && (
          <div className="w-full h-16 inline-flex items-center justify-center border-border border-b">
            There is no one here yet.
          </div>
        )}
      </div>
    </>
  );
};
