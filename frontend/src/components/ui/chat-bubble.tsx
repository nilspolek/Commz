import {
  deleteMessage,
  GATEWAY,
  Message,
  readMessage,
  Summary,
  User,
} from "@/lib/api";
import Markdown from "react-markdown";
import rehypeHighlight from "rehype-highlight";
import remarkGfm from "remark-gfm";
import "highlight.js/styles/atom-one-light.css";
import { format } from "date-fns";
import { cn, getPictureUrl } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import { Check, CheckCheck, Ellipsis, Info, Loader2 } from "lucide-react";
import { memo, useEffect } from "react";
import { Link } from "react-router";
import Draggable from "react-draggable";

const ChatBubbleComponent = ({
  index,
  messages,
  users,
  user,
  setReply,
}: {
  user: User;
  users: User[];
  index: number;
  messages: Message[];
  setReply: (id: Message) => void;
}) => {
  const message = messages[index];
  const showProfilePicture =
    messages.at(index + 1)?.sender != message.sender ||
    index == messages.length - 1;
  const lastMessage =
    messages.at(index - 1)?.sender != message.sender || index == 0;

  users.forEach((user) => {
    message.content = message.content.replace(
      "@" + user.id,
      `**${user.first_name} ${user.last_name}**`
    );
  });

  useEffect(() => {
    if (message.sender != user.id && !message.read) readMessage(message.id);
  }, [message, user.id]);

  const reply: Message | undefined = message.reply_to
    ? messages.find((x) => x.id === message.reply_to)
    : undefined;

  if (message.sender == user.id) {
    // we sent this message
    return (
      <Draggable
        bounds={{ left: -200, right: 0, top: 0, bottom: 0 }}
        position={{ x: 0, y: 0 }}
        onStop={(_, data) => {
          const element = document.getElementById(`message-${message.id}`);
          if (element) {
            element.style.transition = "transform 0.3s ease";
            setTimeout(() => {
              element.style.transition = "";
            }, 300);
          }
          if (Math.abs(data.x) > 100) {
            setReply(message);
          }
        }}
      >
        <div
          id={"message-" + message.id}
          className={cn(
            "ml-auto max-w-[66%] min-w-16 bg-primary bg-opacity-10 text-primary-foreground p-2 rounded-md relative z-10 markdown prose prose-sm prose-invert pb-6",
            lastMessage && "rounded-br-none"
          )}
        >
          {!message.deleted && reply && (
            <div
              className="rounded-lg bg-opacity-20 p-2 bg-black inline-flex gap-1 cursor-pointer overflow-hidden"
              onClick={() => {
                const element = document.getElementById("message-" + reply.id);
                if (element) {
                  element.scrollIntoView({ behavior: "smooth" });
                }
              }}
            >
              <div className="min-h-8 w-1.5 bg-white bg-opacity-50 rounded-full shrink-0"></div>
              {reply.deleted && "Deleted Message"}
              {reply.content && !reply.deleted && reply.content.slice(0, 150)}
              {reply.content &&
                !reply.deleted &&
                reply.content.length > 150 &&
                "..."}
              {!reply.deleted && reply.media && reply.media.length > 0 && (
                <img
                  src={GATEWAY + "media/" + reply.media[0]}
                  className="rounded-lg h-16"
                  draggable={false}
                />
              )}
            </div>
          )}
          {!message.deleted && message.command && (
            <>
              <span className="bg-card text-card-foreground p-1 rounded-sm opacity-50">
                /{message.command}
              </span>{" "}
              {message.content}
            </>
          )}
          {!message.deleted && !message.command && (
            <Markdown
              rehypePlugins={[rehypeHighlight]}
              remarkPlugins={[remarkGfm]}
              className="markdown"
            >
              {message.content}
            </Markdown>
          )}
          {message.deleted && <b className="text-sm">Deleted Message</b>}
          {!message.deleted &&
            message.media?.map((x, i) => (
              <Link to={GATEWAY + "media/" + x} target="_blank">
                <img
                  src={GATEWAY + "media/" + x}
                  key={i}
                  className="rounded-md my-1"
                  draggable={false}
                />
              </Link>
            ))}
          <span className="text-xs text-primary-foreground opacity-50 w-full inline-flex gap-1 items-center justify-end absolute bottom-1 right-2">
            {format(message.timestamp, "HH:mm")}
            {message.loading && <Loader2 className="animate-spin w-3 h-3" />}
            {!message.loading && !message.read && <Check className="w-3 h-3" />}
            {!message.loading && message.read && (
              <CheckCheck className="w-3 h-3" />
            )}
          </span>
          {!message.deleted && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Ellipsis className="absolute text-sm right-2 top-0 hover:opacity-50 opacity-0 transition-all cursor-pointer" />
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuGroup>
                  <DropdownMenuItem onClick={() => deleteMessage(message.id)}>
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </Draggable>
    );
  }

  // get the senders user
  const sender = users.find((x) => x.id == message.sender);

  /// if no sender was found -> we have an AI message
  if (!sender) {
    return (
      <Draggable
        bounds={{ left: 0, right: 200, top: 0, bottom: 0 }}
        position={{ x: 0, y: 0 }}
        onStop={(_, data) => {
          const element = document.getElementById(`message-${message.id}`);
          if (element) {
            element.style.transition = "transform 0.3s ease";
            setTimeout(() => {
              element.style.transition = "";
            }, 300);
          }
          if (Math.abs(data.x) > 100) {
            setReply(message);
          }
        }}
      >
        <div
          className="inline-flex gap-2 max-w-[66%] z-10 markdown"
          id={"message-" + message.id}
        >
          {showProfilePicture && (
            <img
              src={"https://ui-avatars.com/api/?name=AI"}
              alt={"AI"}
              className="w-10 h-10 rounded-full"
              draggable={false}
            />
          )}
          <div
            className={cn(
              "bg-card bg-opacity-10 p-2 min-w-28 rounded-md relative",
              !showProfilePicture && "ml-12", // 10rem with + 2rem gap
              showProfilePicture && "rounded-tl-none"
            )}
          >
            {!message.deleted && (
              <Markdown rehypePlugins={[rehypeHighlight, remarkGfm]}>
                {message.content}
              </Markdown>
            )}
            {message.deleted && <b className="text-sm">Deleted Message</b>}
            <div className="text-xs text-muted-foreground right-2 w-full text-right">
              {format(message.timestamp, "HH:mm")}
            </div>
          </div>
        </div>
      </Draggable>
    );
  }

  return (
    <Draggable
      bounds={{ left: 0, right: 200, top: 0, bottom: 0 }}
      position={{ x: 0, y: 0 }}
      onStop={(_, data) => {
        const element = document.getElementById(`message-${message.id}`);
        if (element) {
          element.style.transition = "transform 0.3s ease";
          setTimeout(() => {
            element.style.transition = "";
          }, 300);
        }
        if (Math.abs(data.x) > 100) {
          setReply(message);
        }
      }}
    >
      <div
        className="inline-flex gap-2 max-w-[66%] z-10 markdown prose prose-sm prose-invert"
        id={"message-" + message.id}
      >
        {showProfilePicture && (
          <img
            src={getPictureUrl(sender)}
            alt={sender.first_name}
            className="w-10 h-10 rounded-full"
            draggable={false}
          />
        )}
        <div
          className={cn(
            "bg-card bg-opacity-10 p-2 min-w-28 rounded-md relative",
            !showProfilePicture && "ml-12", // 10rem with + 2rem gap
            showProfilePicture && "rounded-tl-none"
          )}
        >
          {!message.deleted && reply && (
            <div
              className="rounded-lg bg-opacity-5 p-2 bg-black inline-flex gap-1 cursor-pointer overflow-hidden"
              onClick={() => {
                const element = document.getElementById("message-" + reply.id);
                if (element) {
                  element.scrollIntoView({ behavior: "smooth" });
                }
              }}
            >
              <div className="min-h-8 w-1.5 bg-primary bg-opacity-50 rounded-full shrink-0"></div>
              {reply.deleted && "Deleted Message"}
              {reply.content && !reply.deleted && reply.content.slice(0, 150)}
              {reply.content &&
                !reply.deleted &&
                reply.content.length > 150 &&
                "..."}
              {reply.media && reply.media.length > 0 && (
                <img
                  src={GATEWAY + "media/" + reply.media[0]}
                  className="rounded-lg h-16"
                />
              )}
            </div>
          )}
          {!message.deleted && message.command && (
            <>
              <span className="bg-primary text-primary-foreground p-1 rounded-sm opacity-50">
                /{message.command}
              </span>{" "}
              {message.content}
            </>
          )}
          {!message.deleted && !message.command && (
            <Markdown rehypePlugins={[rehypeHighlight, remarkGfm]}>
              {message.content}
            </Markdown>
          )}
          {message.deleted && <b className="text-sm">Deleted Message</b>}
          {!message.deleted &&
            message.media?.map((x, i) => (
              <Link to={GATEWAY + "media/" + x} target="_blank">
                <img
                  src={GATEWAY + "media/" + x}
                  key={i}
                  className="rounded-md my-1"
                  draggable={false}
                />
              </Link>
            ))}
          <div className="text-xs text-muted-foreground right-2 w-full text-right">
            {format(message.timestamp, "HH:mm")}
          </div>
        </div>
      </div>
    </Draggable>
  );
};

export const SummaryBubble = ({ summary }: { summary: Summary }) => {
  return (
    <div
      className={cn(
        "ml-auto max-w-[66%] bg-primary bg-opacity-10 text-primary-foreground p-2 rounded-md relative pr-12 z-10 markdown prose prose-sm prose-invert"
      )}
    >
      <div className="font-medium opacity-50">Summary: </div>
      <Markdown
        rehypePlugins={[rehypeHighlight]}
        remarkPlugins={[remarkGfm]}
        className="markdown"
      >
        {summary.summary}
      </Markdown>
      <div className="text-xs opacity-50 inline-flex gap-1 items-center">
        <Info className="h-2 w-2" />
        This summary is only visible for you.
      </div>
    </div>
  );
};

export const ChatBubble = memo(ChatBubbleComponent, (prev, next) => {
  return prev == next;
});
