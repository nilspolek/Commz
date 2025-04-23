import { useEffect } from "react";
import { Chat } from "./components/ui/chat";
import { Sidebar } from "./components/ui/sidebar";
import { useAuth } from "./provider/auth-provider";
import { ChatProvider } from "./provider/chat-provider";
import { useNavigate } from "react-router";
import { DialogProvider } from "./provider/dialog-provider";

export default function App() {
  const { loggedIn, loading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!loggedIn && !loading) navigate("/login");
  }, [loggedIn, navigate, loading]);

  return (
    <main>
      <ChatProvider>
        <DialogProvider>
          <div className="w-dvw h-dvh inline-flex">
            <Sidebar />
            <Chat />
          </div>
        </DialogProvider>
      </ChatProvider>
    </main>
  );
}
