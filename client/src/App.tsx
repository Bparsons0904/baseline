import "./App.scss";
import { Component } from "solid-js";
import { AuthProvider } from "./context/AuthContext";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { WebSocketProvider } from "@context/WebSocketProvider";
import { RouteSectionProps } from "@solidjs/router";
import { NavBar } from "@components/layout/Navbar/Navbar";

const App: Component<RouteSectionProps<unknown>> = (props) => {
  const queryClient = new QueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <WebSocketProvider>
          <NavBar />
          <main class="content">{props.children}</main>
        </WebSocketProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
};

export default App;
