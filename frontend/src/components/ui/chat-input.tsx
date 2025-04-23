import {
  Loader2,
  Paperclip,
  Pen,
  PenLine,
  Send,
  Sparkles,
  XIcon,
} from "lucide-react";
import { Button } from "./button";
import { Textarea } from "./textarea";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { Card } from "./card";
import { AxiosError } from "axios";
import { useToast } from "@/hooks/use-toast";
import {
  aiFix,
  aiRewrite,
  GATEWAY,
  getSummary,
  Message,
  Summary,
  uploadImage,
} from "@/lib/api";
import { useChat } from "@/provider/chat-provider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "./dropdown-menu";

export const ChatInput = ({
  submit,
  setSummary,
  reply,
  setReply,
}: {
  submit: (message: string, command?: string, files?: string[]) => void;
  setSummary: React.Dispatch<React.SetStateAction<Summary | undefined>>;
  onImage?: (imageData: File) => void;
  reply?: Message;
  setReply: (id: Message | undefined) => void;
}) => {
  const [message, setMessage] = useState("");
  const { chat } = useChat();

  const showCommand = message.startsWith("/");
  const commands = [
    {
      name: "summary",
      arguments: [],
      description: "Create a summary of the current chat",
    },
    {
      name: "fix",
      arguments: ["message"],
      description: "Checks your current message for spelling mistakes",
    },
    {
      name: "rewrite",
      arguments: ["message"],
      description: "Rewrite your message using AI",
    },
    {
      name: "guess",
      arguments: ["topic"],
      description: "Start a new guess game.",
    },
    {
      name: "guess",
      arguments: ["word"],
      description: "To guess a word in the current game.",
    },
  ];
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();
  const [files, setFiles] = useState<string[]>([]);

  async function selectImage() {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = "image/*";
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;

      const reader = new FileReader();
      reader.readAsArrayBuffer(file);
      reader.onload = async () => {
        try {
          setLoading(true);
          const bytes = new Uint8Array(reader.result as ArrayBuffer);
          // You can now use the bytes array for further processing
          const result = await uploadImage(bytes, file.type);
          setFiles((prev) => [...prev, result]);
        } catch {
          toast({
            title: "Upload failed",
            description: "Make sure to select an image",
            variant: "destructive",
          });
        } finally {
          setLoading(false);
        }
      };
    };
    input.click();
  }

  const executeCommand = async (command: {
    name: string;
    description: string;
  }) => {
    const commandContent = message.replace(`/${command.name} `, "");
    let count = 0;
    try {
      setLoading(true);
      switch (command.name) {
        case "summary":
          if (!chat) {
            setLoading(false);
            return;
          }
          setMessage("");
          setSummary(undefined);
          await getSummary(chat, (chunk) => {
            setLoading(false);

            if (chunk.done_reason === "error") {
              toast({
                title: "Failed",
                description: chunk.response,
                variant: "destructive",
              });
              return;
            }

            setSummary((prev) => {
              if (prev?.chat_id != chat.id)
                return { chat_id: chat.id!, summary: chunk.response };
              const summary: Summary =
                prev == undefined
                  ? { chat_id: chat.id!, summary: chunk.response }
                  : { ...prev, summary: prev.summary + chunk.response };
              return summary;
            });
          });
          break;
        case "guess":
          if (!commandContent) {
            toast({
              title: "Failed",
              description: "Please provide a topic",
              variant: "destructive",
            });
            setLoading(false);
            return;
          }
          setMessage("");
          submit(commandContent, command.name);
          setLoading(false);
          break;
        case "rewrite":
          if (!commandContent) {
            setLoading(false);
            return;
          }
          await aiRewrite(commandContent, (chunk) => {
            if (count == 0) setMessage("");
            count++;
            if (chunk.done) setLoading(false);
            if (chunk.done_reason === "error") {
              toast({
                title: "Failed",
                description: chunk.response,
                variant: "destructive",
              });
              return;
            }
            setMessage((prev) =>
              prev == undefined ? chunk.response : prev + chunk.response
            );
          });
          break;
        case "fix":
          if (!commandContent) {
            setLoading(false);
            return;
          }
          await aiFix(commandContent, (chunk) => {
            if (count == 0) setMessage("");
            count++;
            if (chunk.done) setLoading(false);
            if (chunk.done_reason === "error") {
              toast({
                title: "Failed",
                description: chunk.response,
                variant: "destructive",
              });
              return;
            }
            setMessage((prev) =>
              prev == undefined ? chunk.response : prev + chunk.response
            );
          });
          break;
      }
    } catch (e) {
      setLoading(false);
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

  const handlePaste = async (e: React.ClipboardEvent) => {
    try {
      setLoading(true);
      const items = Array.from(e.clipboardData.items);
      for (const item of items) {
        if (item.type.startsWith("image/")) {
          e.preventDefault();
          const file = item.getAsFile();
          if (!file) return;
          const bytes = await file.arrayBuffer();
          const byteArray = new Uint8Array(bytes);

          const image = await uploadImage(byteArray, file.type);
          setFiles((prev) => [...prev, image]);
        }
      }
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter") {
      if (!e.shiftKey) {
        e.preventDefault();
        if (message.trim() == "" && files.length == 0) return;
        if (showCommand && filteredCommands.length > 0) {
          executeCommand(filteredCommands[0]);
          return;
        }
        submit(message, undefined, files);
        setMessage("");
        setFiles([]);
      }
    }
  };

  const filteredCommands = commands.filter((command) =>
    command.name.startsWith(message.replace("/", "").split(" ")[0])
  );

  // Calculate height based on number of filtered commands
  const commandHeight = 40; // height of each command button in pixels
  const style = {
    "--command-height": `${filteredCommands.length * commandHeight + 16}px`,
  } as React.CSSProperties;

  return (
    <>
      <div className="w-full relative">
        <div className="w-full px-4 absolute bottom-full -mb-3 z-10 flex gap-2">
          {files.map((x, i) => (
            <div
              key={i}
              className="w-16 h-16 rounded-sm border-border border shadow-sm relative"
            >
              <img
                src={GATEWAY + "media/" + x}
                className="w-full h-full object-cover rounded-sm"
              />
              <XIcon
                className="w-4 h-4 absolute -top-2 -right-2 bg-border rounded-full text-muted-foreground hover:scale-110 transition-all cursor-pointer"
                onClick={() => {
                  files.splice(i, 1);
                  setFiles([...files]);
                }}
              />
            </div>
          ))}
        </div>
        <Card
          style={style}
          className={cn(
            "absolute bottom-full -mb-5 left-5 z-10 w-[60%] bg-secondary p-2 transition-all duration-300 text-muted-foreground flex gap-2 items-center flex-wrap",
            reply ? "opacity-100 translate-y-0" : "opacity-0 translate-y-5"
          )}
        >
          <span className="text-primary inline-flex gap-2 items-center mr-2 w-full shrink-0">
            Reply to:
            <XIcon
              className="bg-border rounded-full text-muted-foreground p-1 cursor-pointer ml-auto"
              onClick={() => setReply(undefined)}
            />
          </span>
          {reply && reply.content && reply.content.slice(0, 50)}
          {reply && reply.content && reply.content.length > 50 && "..."}
          {reply &&
            reply.media &&
            reply.media.map((x, i) => (
              <img
                key={i}
                src={GATEWAY + "media/" + x}
                className="w-8 h-8 rounded-md border-border border"
              />
            ))}
        </Card>
        <Card
          style={style}
          className={cn(
            "absolute bottom-full -mb-5 left-5 z-10 w-[60%] bg-secondary",
            "transition-all duration-300",
            showCommand && filteredCommands.length > 0
              ? "opacity-100 translate-y-0"
              : "opacity-0 translate-y-5"
          )}
        >
          <div
            className={cn(
              "grid transition-[grid-template-rows] duration-300",
              showCommand
                ? "grid-rows-[var(--command-height)]"
                : "grid-rows-[0fr]"
            )}
          >
            <div className="overflow-hidden">
              <div className="p-2">
                {filteredCommands.map((command, i) => (
                  <div
                    className={cn(
                      "transition-all duration-300",
                      showCommand ? "opacity-100" : "opacity-0"
                    )}
                    style={{
                      transitionDelay: i * 150 + "ms",
                    }}
                  >
                    <Button
                      key={command.name}
                      className={cn(
                        "flex gap-2 w-full justify-start hover:bg-border transition-all duration-300 h-[40px]"
                      )}
                      variant={"ghost"}
                      onClick={() => {
                        // check if the command is already filled out
                        if (message.trim().startsWith(command.name)) {
                          executeCommand(command);
                          return;
                        }
                        setMessage("/" + command.name + " ");
                      }}
                    >
                      <span className="text-primary">{command.name}</span>
                      {command.arguments.length > 0 && (
                        <span className="text-muted-foreground">
                          {command.arguments.map((x) => `[${x}]`).join(" ")}
                        </span>
                      )}
                      <span>-</span>
                      <span className="text-muted-foreground">
                        {command.description}
                      </span>
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Card>
        <form
          className="mt-auto p-4 inline-flex gap-2 w-full relative z-20"
          onSubmit={(e) => {
            e.preventDefault();
            if (message.trim() == "" && files.length == 0) return;
            if (showCommand && filteredCommands.length > 0) {
              executeCommand(filteredCommands[0]);
              return;
            }
            submit(message, undefined, files);
            setMessage("");
            setFiles([]);
          }}
        >
          <div className="relative flex-1">
            <Textarea
              placeholder="Message"
              disabled={loading}
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyDown={handleKeyDown}
              onPaste={handlePaste}
              rows={Math.min(message.split("\n").length, 5)}
              className={cn(
                "transition-all duration-500 disabled:opacity-100 opacity-100 disabled:text-muted-foreground bg-white pr-12 resize-none min-h-[40px] max-h-[120px]"
              )}
              onDrop={async (e) => {
                e.preventDefault();
                e.stopPropagation();
                try {
                  setLoading(true);
                  const items = Array.from(e.dataTransfer.files);
                  for (const file of items) {
                    if (file.type.startsWith("image/")) {
                      const bytes = await file.arrayBuffer();
                      const byteArray = new Uint8Array(bytes);
                      const image = await uploadImage(byteArray, file.type);
                      setFiles((prev) => [...prev, image]);
                    }
                  }
                } finally {
                  setLoading(false);
                }
              }}
            />
            <Loader2
              className={cn(
                "absolute right-1.5 top-1.5 text-muted-foreground animate-spin transition-all",
                loading ? "opacity-100" : "opacity-0"
              )}
            />

            <div
              className={cn(
                "absolute inset-0 pointer-events-none opacity-0 transition-all",
                loading && "opacity-100"
              )}
            >
              <div className="h-full w-full animate-pulse bg-primary/10 rounded-md" />
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant={"ghost"}
                  className={cn("absolute right-0 top-0", loading && "hidden")}
                >
                  <Sparkles className="text-primary" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuLabel>Options</DropdownMenuLabel>
                <DropdownMenuItem
                  onClick={() =>
                    executeCommand({
                      name: "fix",
                      description: "",
                    })
                  }
                >
                  <Pen className="w-4 h-4 mr-2" /> Fix
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() =>
                    executeCommand({
                      name: "rewrite",
                      description: "",
                    })
                  }
                >
                  <PenLine className="w-4 h-4 mr-2" /> Rewrite
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() =>
                    executeCommand({
                      name: "summary",
                      description: "",
                    })
                  }
                >
                  <Sparkles className="w-4 h-4 mr-2" /> Summary
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          {/* TODO: open finder and make a selection to images */}
          <Button
            disabled={loading}
            onClick={(e) => {
              e.preventDefault();
              selectImage();
            }}
          >
            <Paperclip />
          </Button>
          <Button disabled={loading}>
            <Send />
          </Button>
        </form>
      </div>
    </>
  );
};
