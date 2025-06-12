import { Route } from "@solidjs/router";
import { Component, lazy } from "solid-js";
import {
  ProtectedRoute,
  ProtectedRouteType,
} from "@components/common/ProtectedRoute/ProtectedRoute";

const HomePage = lazy(() => import("@pages/Home/Home"));
const LoginPage = lazy(() => import("@pages/Auth/Login"));
const ProfilePage = lazy(() => import("@pages/Profile/Profile"));

export const Routes: Component = () => {
  return (
    <>
      <Route path="/" component={HomePage} />
      <Route path="/login" component={LoginPage} />
      <Route
        path="/profile"
        component={() => (
          <ProtectedRoute type={ProtectedRouteType.User}>
            <ProfilePage />
          </ProtectedRoute>
        )}
      />
    </>
  );
};
