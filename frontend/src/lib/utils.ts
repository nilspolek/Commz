import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { Chat, GATEWAY, User } from "./api";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function getChatMember(chat: Chat, self: User, users: User[]) {
  const otherUserId = chat.members.filter((x) => x != self.id).at(0);
  const otherUser = users.find((x) => x.id == otherUserId);
  if (!otherUser || !otherUserId) {
    return undefined;
  }
  return otherUser;
}

export function getPictureUrl(user: User | string) {
  if (typeof user == "object") {
    if (user.picture) {
      return GATEWAY + "media/" + user.picture;
    }
    return "https://ui-avatars.com/api/?name=" + user.first_name;
  }

  return GATEWAY + "media/" + user;
}
