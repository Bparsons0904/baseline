import { Component, JSX, createEffect } from "solid-js";
import { useAuth } from "@context/AuthContext";
import { useNavigate } from "@solidjs/router";

export enum ProtectedRouteType {
  User = 0,
  Admin = 1,
}

interface ProtectedRouteProps {
  type?: ProtectedRouteType;
  children: JSX.Element;
}

export const ProtectedRoute: Component<ProtectedRouteProps> = (props) => {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  createEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login", { replace: true });
    }
  });

  return <>{isAuthenticated() ? props.children : null}</>;
};
