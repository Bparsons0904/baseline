import { env } from "@services/env.service";
import axios from "axios";

export const apiClient = axios.create({
  baseURL: env.apiUrl + "/api/",
  timeout: 10000,
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Client-Type": "solid",
  },
  withCredentials: true,
});

export const initializeTokenInterceptor = (
  setToken: (token: string) => void,
) => {
  apiClient.interceptors.response.clear();
  apiClient.interceptors.response.use(
    (response) => {
      const token = response.headers["x-auth-token"];
      if (token) {
        setToken(token);
      }
      return response;
    },
    (error) => {
      return Promise.reject(error);
    },
  );
};

export const getApi = async <T>(
  url: string,
  params?: Record<string, string>,
): Promise<T> => {
  const response = await apiClient.get(`/${url}`, { params });

  if (response.data.error) {
    console.error(response.data.error);
    throw new Error(response.data.error.message);
  }

  return response.data;
};

export const postApi = async <T, U>(url: string, data: U): Promise<T> => {
  const response = await apiClient.post(`/${url}`, data);

  if (response.data.error) {
    console.error(response.data.error);
    throw new Error(response.data.error.message);
  }

  return response.data;
};
