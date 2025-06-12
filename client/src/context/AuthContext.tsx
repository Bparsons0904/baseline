import {
  getApi,
  initializeTokenInterceptor,
  postApi,
} from "@services/api/api.service";
// import { useLogin } from "@services/api/post";
import { useNavigate } from "@solidjs/router";
import { useMutation, useQuery } from "@tanstack/solid-query";
import {
  createContext,
  useContext,
  createSignal,
  JSX,
  Accessor,
  createEffect,
} from "solid-js";
import { createStore } from "solid-js/store";
import { User } from "src/types/User";

type AuthContextValue = {
  isAuthenticated: Accessor<boolean>;
  user: User | null;
  authToken: Accessor<string | null>;
  // setAuthToken: (token: string | null) => void;
  login: (credentials: { login: string; password: string }) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue>({} as AuthContextValue);

export function AuthProvider(props: { children: JSX.Element }) {
  const navigate = useNavigate();
  const [user, setUser] = createStore(null);
  const [isAuthenticated, setIsAuthenticated] = createSignal(false);
  const [authToken, setAuthToken] = createSignal<string | null>(null);

  initializeTokenInterceptor(setAuthToken);

  const getUserResponse = useQuery(() => ({
    queryKey: ["user"],
    queryFn: () => getApi<{ user: User }>("users"),
    refetchOnWindowFocus: false,
    retry: false,
  }));

  createEffect(() => {
    if (getUserResponse.isSuccess && getUserResponse.data.user) {
      setUser(getUserResponse.data.user);
      setIsAuthenticated(true);
      navigate("/");
    }
  });

  interface LoginCredentials {
    login: string;
    password: string;
  }

  const loginUser = useMutation(() => ({
    mutationFn: (credentials: LoginCredentials) =>
      postApi<User, LoginCredentials>("users/login", credentials),
    // onSuccess: () => {
    //   navigate("/");
    // },
  }));
  const login = async (credentials: { login: string; password: string }) => {
    const user = await loginUser.mutateAsync(credentials);
    if (!user) return;
    setUser(user);
    setIsAuthenticated(!!user);
    navigate("/");
  };

  const logoutUser = useMutation(() => ({
    mutationFn: () => postApi("users/logout", {}),
    onSuccess: () => {
      setUser(null);
      setIsAuthenticated(false);
      setAuthToken(null);
      navigate("/login");
    },
  }));

  const logout = async () => {
    logoutUser.mutate();
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        login,
        logout,
        authToken,
        // setAuthToken,
      }}
    >
      {props.children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
