import {
  createContext,
  useContext,
  createSignal,
  createEffect,
  JSX,
  onCleanup,
} from "solid-js";
import { env } from "@services/env.service";
import { useAuth } from "./AuthContext";

interface WebSocketContextValue {
  isConnected: () => boolean;
  connectionState: () => number; // WebSocket.CONNECTING, OPEN, CLOSING, CLOSED
  sendMessage: (message: string) => void;
  lastMessage: () => string | null;
}

const WebSocketContext = createContext<WebSocketContextValue>(
  {} as WebSocketContextValue,
);

export function WebSocketProvider(props: { children: JSX.Element }) {
  const [socket, setSocket] = createSignal<WebSocket | null>(null);
  const [connectionState, setConnectionState] = createSignal<number>(
    WebSocket.CLOSED,
  );
  const [lastMessage, setLastMessage] = createSignal<string | null>(null);

  const { isAuthenticated, authToken } = useAuth();

  createEffect(() => {
    if (!isAuthenticated() || !authToken()) return () => {};
    console.log("WebSocketProvider: Creating new connection");
    console.log("token:", authToken());

    let wsUrl = env.wsUrl || "ws://localhost:8280/ws";
    console.log("WebSocket URL:", wsUrl);

    wsUrl += `?token=${authToken()}`;

    try {
      const newSocket = new WebSocket(wsUrl);
      console.log("WebSocket created:", newSocket);
      setConnectionState(WebSocket.CONNECTING);

      newSocket.addEventListener("open", () => {
        console.log("WebSocket connection established");
        setConnectionState(WebSocket.OPEN);

        // newSocket.send("Hello from client!");
      });

      newSocket.addEventListener("message", (event) => {
        console.log("WebSocket message received:", event.data);
        setLastMessage(event.data);
      });

      newSocket.addEventListener("close", (event) => {
        console.log("WebSocket connection closed:", event);
        setConnectionState(WebSocket.CLOSED);
      });

      newSocket.addEventListener("error", (error) => {
        console.error("WebSocket error:", error);
        setConnectionState(WebSocket.CLOSED);
      });

      setSocket(newSocket);

      onCleanup(() => {
        console.log("Cleaning up WebSocket connection");
        if (newSocket.readyState === WebSocket.OPEN) {
          newSocket.close();
        }
      });
    } catch (error) {
      console.error("Error creating WebSocket:", error);
      return () => {};
    }
  });

  const isConnected = () => connectionState() === WebSocket.OPEN;

  const sendMessage = (message: string) => {
    const currentSocket = socket();
    if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
      currentSocket.send(message);
    } else {
      console.error("Cannot send message: WebSocket not connected");
    }
  };

  const contextValue: WebSocketContextValue = {
    isConnected,
    connectionState,
    sendMessage,
    lastMessage,
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {props.children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  return useContext(WebSocketContext);
}
