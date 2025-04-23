import { Chat, Message } from "@/lib/api";
import React, { useContext, useEffect, useState } from "react";
import * as api from "@/lib/api";
import { useAuth } from "./auth-provider";
import { useToast } from "@/hooks/use-toast";

interface ChatContextType {
  loading: boolean;
  chats: Chat[];
  chat?: Chat;
  setChat: (chat: Chat) => void;
  newChat: (user_id: string, message: string) => Promise<Chat>;
  sendMessage: (
    chat_id: string,
    message: string,
    command?: string,
    media?: string[],
    reply_to?: string
  ) => Promise<Message>;
  updateMessage: (message: Message) => Promise<void>;
  deleteMessage: (message: Message) => Promise<void>;
  refresh: () => void;
}

const ChatContext = React.createContext<ChatContextType | null>(null);

export const ChatProvider = ({ children }: { children: React.ReactNode }) => {
  const [chats, setChats] = useState<Chat[]>([]);
  const [chat, setChat] = useState<Chat | undefined>();
  const { loggedIn, user, users } = useAuth();
  const [socket, setSocket] = useState<WebSocket>();
  const [isConnecting, setIsConnecting] = useState(false);
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  const connectWebSocket = () => {
    if (socket?.readyState === WebSocket.OPEN || isConnecting || !loggedIn)
      return;

    // Close existing socket if it exists
    if (socket) {
      socket.close();
    }

    setIsConnecting(true);
    let wsUrl = `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${
      window.location.host
    }/api/ws`;
    if (import.meta.env.DEV) wsUrl = "ws://localhost:4242/ws";
    const s = new WebSocket(wsUrl);

    s.onopen = () => {
      setSocket(s);
      setIsConnecting(false);
    };

    s.onmessage = (event) => {
      const message: Message | { error: string } = JSON.parse(event.data);
      if ("error" in message) {
        toast({
          title: "Error",
          description: message.error,
          variant: "destructive",
        });
      } else
        setChats((prev) => {
          const chat = prev.find((x) => x.id === message.chat_id);
          if (!chat) {
            // if there is no chat, we might need to refetch, to get the new created chat
            api.getChats().then((c) => setChats(c));
            return prev;
          }

          // Check if message already exists to prevent duplicates
          if (
            chat.messages?.some(
              (m) => m.id === message.id && message.id !== "-1"
            )
          ) {
            const index = chat.messages.findIndex((x) => x.id == message.id);
            chat.messages[index] = message;

            return prev.map((c) => {
              if (c.id === chat.id) {
                return chat;
              }
              return c;
            });
          }

          return prev.map((c) => {
            if (c.id === message.chat_id) {
              return {
                ...c,
                messages: [...(c.messages || []), { ...message }],
              };
            }
            return c;
          });
        });
    };

    s.onclose = () => {
      setSocket(undefined);
      setIsConnecting(false);
      // Attempt to reconnect after 3 seconds
      setTimeout(connectWebSocket, 3000);
    };

    s.onerror = () => {
      s.close();
    };
  };

  useEffect(() => {
    connectWebSocket();
    return () => {
      if (socket) {
        socket.close();
      }
    };
  }, [loggedIn]);

  useEffect(() => {
    const load = async () => {
      try {
        setChats(await api.getChats());
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [loggedIn]);

  useEffect(() => {
    if (!chat) {
      // const update = chats.at(0);
      // setChat(update);
      return;
    }
    const selected = chats.find((x) => x.id == chat?.id);
    if (selected && selected.messages?.length != chat?.messages?.length)
      setChat(selected);
  }, [chats, chat, users, user]);

  const refresh = async () => {
    api.getChats().then((x) => {
      setChats(x);
      if (chat && chat.id) setChat(x.find((x) => x.id == chat.id));
    });
  };

  const updateMessage = async (message: Message) => {
    socket!.send(
      JSON.stringify({
        content: message.content,
        chat_id: message.chat_id,
        id: message.id,
      })
    );
  };

  const deleteMessage = async (message: Message) => {
    socket!.send(
      JSON.stringify({
        content: message.content,
        chat_id: message.chat_id,
        id: message.id,
        deleted: true,
      })
    );
  };

  return (
    <ChatContext.Provider
      value={{
        loading,
        chats,
        chat,
        setChat,
        newChat: api.createDirectChat,
        sendMessage: api.sendMessage,
        updateMessage,
        deleteMessage,
        refresh,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
};

export const useChat = () => {
  const chat = useContext(ChatContext);
  if (!chat) throw new Error("can onyl useChat in ChatProvider");
  return chat;
};
