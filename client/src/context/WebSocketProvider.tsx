import {
  createContext,
  useContext,
  createSignal,
  createEffect,
  onCleanup,
  JSX,
} from "solid-js";
import {
  createReconnectingWS,
  createWSState,
} from "@solid-primitives/websocket";
import { env } from "@services/env.service";
import { useAuth } from "./AuthContext";

export const MessageType = {
  PING: "ping",
  PONG: "pong",
  MESSAGE: "message",
  BROADCAST: "broadcast",
  ERROR: "error",
  USER_JOIN: "user_join",
  USER_LEAVE: "user_leave",
  AUTH_REQUEST: "auth_request",
  AUTH_RESPONSE: "auth_response",
  AUTH_SUCCESS: "auth_success",
  AUTH_FAILURE: "auth_failure",
} as const;

export type ChannelType = "system" | "user";

export interface WebSocketMessage {
  id: string;
  type: string;
  channel: ChannelType;
  action: string;
  userId?: string;
  data?: Record<string, unknown>;
  timestamp: string;
}

export enum ConnectionState {
  Connecting = "connecting",
  Connected = "connected",
  Authenticating = "authenticating",
  Authenticated = "authenticated",
  Disconnecting = "disconnecting",
  Disconnected = "disconnected",
  Failed = "failed",
}

interface WebSocketContextValue {
  connectionState: () => ConnectionState;
  isConnected: () => boolean;
  isAuthenticated: () => boolean;
  lastError: () => string | null;
  lastMessage: () => string;
  sendMessage: (message: string) => void;
  reconnect: () => void;
}

const WebSocketContext = createContext<WebSocketContextValue>(
  {} as WebSocketContextValue,
);

interface WebSocketProviderProps {
  children: JSX.Element;
  debug?: boolean;
}

export function WebSocketProvider(props: WebSocketProviderProps) {
  const { isAuthenticated, authToken } = useAuth();
  // const debug = props.debug ?? import.meta.env.DEV;

  const [lastError, setLastError] = createSignal<string | null>(null);
  const [wsInstance, setWsInstance] = createSignal<ReturnType<
    typeof createReconnectingWS
  > | null>(null);
  const [wsAuthenticated, setWsAuthenticated] = createSignal<boolean>(false);

  const log = (message: string, ...args: unknown[]) => {
    // if (!debug) {
    console.log(`[WebSocket] ${message}`, ...args);
    // }
  };

  const getWebSocketUrl = () => {
    if (!isAuthenticated() || !authToken()) {
      return null;
    }
    return env.wsUrl;
  };

  const handleAuthRequest = () => {
    log("Handling auth request");
    const token = authToken();

    if (!token) {
      log("No auth token available");
      setLastError("No authentication token available");
      return;
    }

    const authResponse: WebSocketMessage = {
      id: crypto.randomUUID(),
      type: MessageType.AUTH_RESPONSE,
      channel: "system",
      action: "authenticate",
      data: { token },
      timestamp: new Date().toISOString(),
    };

    const ws = wsInstance();
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(authResponse));
      log("Auth response sent");
    }
  };

  const handleMessage = (event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      log("Received message:", message);

      switch (message.type) {
        case MessageType.AUTH_REQUEST:
          handleAuthRequest();
          break;

        case MessageType.AUTH_SUCCESS:
          log("Authentication successful");
          setWsAuthenticated(true);
          setLastError(null);
          break;

        case MessageType.AUTH_FAILURE:
          log("Authentication failed:", message.data?.reason);
          setWsAuthenticated(false);
          setLastError(
            typeof message.data?.reason === "string"
              ? message.data.reason
              : "Authentication failed",
          );
          break;

        default:
          // TODO: Handle other message types
          break;
      }
    } catch (error) {
      log("Failed to parse message:", error);
    }
  };

  createEffect(() => {
    const url = getWebSocketUrl();

    if (!url) {
      log("No URL available, clearing WebSocket");
      setWsInstance(null);
      setWsAuthenticated(false);
      setLastError("Authentication required");
      return;
    }

    log("Creating WebSocket");
    setWsAuthenticated(false);

    try {
      const ws = createReconnectingWS(url);

      // Set up event listeners
      ws.addEventListener("open", () => {
        log("WebSocket connected, waiting for auth request");
        setLastError(null);
      });

      ws.addEventListener("message", handleMessage);

      ws.addEventListener("error", (event) => {
        log("WebSocket error:", event);
        setLastError("Connection error occurred");
      });

      ws.addEventListener("close", (event) => {
        log("WebSocket closed:", event.code, event.reason);
        setWsAuthenticated(false);
        if (event.code !== 1000) {
          // Not normal closure
          setLastError(
            `Connection closed unexpectedly: ${event.reason || "Unknown reason"}`,
          );
        }
      });

      setWsInstance(ws);
      setLastError(null);
    } catch (error) {
      log("Failed to create WebSocket:", error);
      setLastError(
        error instanceof Error ? error.message : "Failed to create connection",
      );
      setWsInstance(null);
    }
  });

  const wsState = () => {
    const ws = wsInstance();
    return ws ? createWSState(ws)() : WebSocket.CLOSED;
  };

  const connectionState = (): ConnectionState => {
    if (!wsInstance()) {
      return ConnectionState.Disconnected;
    }

    const rawState = wsState();

    switch (rawState) {
      case WebSocket.CONNECTING:
        return ConnectionState.Connecting;
      case WebSocket.OPEN:
        return wsAuthenticated()
          ? ConnectionState.Authenticated
          : ConnectionState.Authenticating;
      case WebSocket.CLOSING:
        return ConnectionState.Disconnecting;
      case WebSocket.CLOSED:
        return ConnectionState.Disconnected;
      default:
        return ConnectionState.Failed;
    }
  };

  const isConnected = () => {
    const state = connectionState();
    return (
      state === ConnectionState.Connected ||
      state === ConnectionState.Authenticating ||
      state === ConnectionState.Authenticated
    );
  };

  const isWebSocketAuthenticated = () => wsAuthenticated();

  const lastMessage = () => {
    return "";
  };

  const sendMessage = (message: string) => {
    const ws = wsInstance();

    if (!ws) {
      log("Cannot send message: No WebSocket instance");
      setLastError("Cannot send message: not connected");
      return;
    }

    if (!wsAuthenticated()) {
      log("Cannot send message: WebSocket not authenticated");
      setLastError("Cannot send message: not authenticated");
      return;
    }

    if (wsState() !== WebSocket.OPEN) {
      log("Cannot send message: WebSocket not open");
      setLastError("Cannot send message: connection not ready");
      return;
    }

    try {
      ws.send(message);
      log("Message sent:", message);
    } catch (error) {
      log("Failed to send message:", error);
      setLastError(
        error instanceof Error ? error.message : "Failed to send message",
      );
    }
  };

  const reconnect = () => {
    const ws = wsInstance();
    if (ws && "reconnect" in ws && typeof ws.reconnect === "function") {
      log("Manually triggering reconnection");
      setWsAuthenticated(false);
      ws.reconnect();
    } else {
      log("No reconnect method available, recreating connection");
      setWsInstance(null);
      setWsAuthenticated(false);
    }
  };

  onCleanup(() => {
    log("Cleaning up WebSocket connection");
    const ws = wsInstance();
    if (ws) {
      ws.close(1000, "Component cleanup");
    }
  });

  createEffect(() => {
    const handleBeforeUnload = () => {
      log("Page unloading, closing WebSocket");
      const ws = wsInstance();
      if (ws) {
        ws.close(1000, "Page unload");
      }
    };

    window.addEventListener("beforeunload", handleBeforeUnload);

    onCleanup(() => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
    });
  });

  const contextValue: WebSocketContextValue = {
    connectionState,
    isConnected,
    isAuthenticated: isWebSocketAuthenticated,
    lastError,
    lastMessage,
    sendMessage,
    reconnect,
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {props.children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within WebSocketProvider");
  }
  return context;
}
