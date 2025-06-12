import { Component } from "solid-js";
import styles from "./Auth.module.scss";
import { TextInput } from "@components/common/forms/TextInput/TextInput";
import { createStore } from "solid-js/store";
import { Button } from "@components/common/Button/Button";
import { useAuth } from "@context/AuthContext";

const Login: Component = () => {
  const [loginState, setLoginState] = createStore<{
    login: string;
    password: string;
  }>({
    login: "deadstyle",
    password: "password",
  });

  const { login } = useAuth();

  const handleUpdate = (field: "login" | "password", value: string) => {
    setLoginState(field, value);
  };

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    login(loginState);
  };

  return (
    <div class={styles.auth}>
      <TextInput
        label="Login"
        autoComplete="username"
        onBlur={(value) => handleUpdate("login", value)}
      />
      <TextInput
        label="Password"
        type="password"
        autoComplete="current-password"
        onBlur={(value) => handleUpdate("password", value)}
      />
      <Button type="submit" onClick={handleSubmit}>
        Login
      </Button>
    </div>
  );
};

export default Login;
