import { useAuth } from "@context/AuthContext";
import { Component, Show } from "solid-js";

const Home: Component = () => {
  const { user, isAuthenticated } = useAuth();

  return (
    <div>
      <h1>Welcome to Baseline App</h1>
      <Show 
        when={isAuthenticated()} 
        fallback={<p>Please log in to access the application.</p>}
      >
        <p>Hello, {user?.firstName || 'User'}! You are successfully logged in.</p>
        <div>
          <h2>Dashboard</h2>
          <p>This is your main dashboard. Add your application features here.</p>
        </div>
      </Show>
    </div>
  );
};

export default Home;
