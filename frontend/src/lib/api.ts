import axios from "axios";

axios.defaults.withCredentials = true;
export const GATEWAY = import.meta.env.DEV ? "http://localhost:4242/" : "/api/";

let wsUrl = `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${
  window.location.host
}/api/ai`;
if (import.meta.env.DEV) wsUrl = "ws://localhost:4242/ai";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

export interface User {
  email: string;
  id: string;
  first_name: string;
  last_name: string;
  picture?: string;
}

export interface Chat {
  members: string[];
  name: string;
  id?: string; // if id is null => this is a temporary chat and we have to create it first when we want to send a new message
  createdAt: Date;
  messages?: Message[];
  last_active: Date;
}

export interface Message {
  id: string;
  content: string;
  sender: string;
  chat_id: string;
  timestamp: string;
  command?: string;
  deleted?: boolean;
  loading?: boolean;
  media?: string[];
  reply_to?: string;
  read: boolean;
}

export interface CreateChat {
  name: string;
  members: string[];
}

export interface Summary {
  summary: string;
  chat_id: string;
}

export interface Chunk {
  done: boolean;
  response: string;
  done_reason?: string;
}

export async function login(request: LoginRequest) {
  const { data } = await axios.post(`${GATEWAY}auth/login`, request);
  return data as User;
}

export async function register(request: RegisterRequest) {
  const { data } = await axios.post(`${GATEWAY}auth/register`, request);
  return data as User;
}

export async function getUsers() {
  const { data } = await axios.get(`${GATEWAY}auth/users`);
  return data as User[];
}

export async function getUser() {
  const { data } = await axios.get(`${GATEWAY}auth/user`);
  return data as User;
}

export async function getChats() {
  const { data } = await axios.get(`${GATEWAY}chat`);
  return data as Chat[];
}

export async function sendMessage(
  chat_id: string,
  message: string,
  command?: string,
  media?: string[],
  reply_to?: string
) {
  const { data } = await axios.post(`${GATEWAY}chat/${chat_id}/messages`, {
    message,
    command,
    media,
    reply_to,
  });
  return { ...data, loading: true } as Message;
}

export async function readMessage(message_id: string) {
  const { data } = await axios.get(
    `${GATEWAY}chat/messages/${message_id}/read`
  );
  return { ...data } as Message;
}

export async function deleteMessage(message_id: string) {
  const { data } = await axios.delete(`${GATEWAY}chat/messages/${message_id}`);
  return { ...data, loading: true } as Message;
}

export async function createDirectChat(user_id: string, message: string) {
  const { data } = await axios.post(`${GATEWAY}chat/direct-chat`, {
    message,
    receiver: user_id,
  });
  return data as Chat;
}

export async function createGroupChat(name: string, members: string[]) {
  const { data } = await axios.post(`${GATEWAY}chat/`, {
    name,
    members,
  });
  return data as Chat;
}

export async function getMessages(chatId: string, limit = 30, offset = 0) {
  const { data } = await axios.get(
    `${GATEWAY}chat/${chatId}/messages?limit=${limit}&offset=${offset}`
  );
  return data as Message[];
}

export async function deleteChat(chatId: string) {
  await axios.delete(`${GATEWAY}chat/${chatId}`);
}

export async function updateGroupChat(
  chatId: string,
  name: string,
  members: string[]
) {
  const { data } = await axios.put(`${GATEWAY}chat/${chatId}`, {
    name,
    members,
  });
  return data as Chat;
}

export async function uploadImage(bytes: Uint8Array, contentType: string) {
  const { data } = await axios.post(`${GATEWAY}media`, bytes, {
    headers: { "Content-Type": contentType },
  });
  return data.name as string;
}

export async function updateProfile(user: User) {
  const { data } = await axios.put(`${GATEWAY}auth/user`, {
    ...user,
  });
  return data as User;
}

export async function updatePassword(
  current_password: string,
  new_password: string
) {
  const { data } = await axios.put(`${GATEWAY}auth/user/password`, {
    current_password,
    new_password,
  });
  return data as User;
}

export async function getSummary(
  chat: Chat,
  onReceive: (chunk: Chunk) => void
) {
  const socket = new WebSocket(`${wsUrl}/summarization`);

  socket.onopen = () => {
    socket.send(JSON.stringify(chat));
  };

  socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onReceive(data);
  };

  socket.onerror = () => {
    onReceive({
      done: true,
      response: "Connection failed",
      done_reason: "error",
    });
  };
}

export async function aiRewrite(
  message: string,
  onReceive: (chunk: Chunk) => void
) {
  const socket = new WebSocket(`${wsUrl}/rewrite`);

  socket.onopen = () => {
    socket.send(JSON.stringify({ text: message }));
  };

  socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onReceive(data);
  };

  socket.onerror = () => {
    onReceive({
      done: true,
      response: "Connection failed",
      done_reason: "error",
    });
  };
}

export async function aiFix(
  message: string,
  onReceive: (chunk: Chunk) => void
) {
  const socket = new WebSocket(`${wsUrl}/fix`);

  socket.onopen = () => {
    socket.send(JSON.stringify({ text: message }));
  };

  socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onReceive(data);
  };

  socket.onerror = () => {
    onReceive({
      done: true,
      response: "Connection failed",
      done_reason: "error",
    });
  };
}

export async function userLogout() {
  await axios.get(GATEWAY + "/auth/logout");
}

export async function getVersion(service: string) {
  if (service == "gateway") {
    const { data } = await axios.get(GATEWAY + "/version");
    return data as string;
  }
  const { data } = await axios.get(GATEWAY + service + "/version");
  return data as string;
}
