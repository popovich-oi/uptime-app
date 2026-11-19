import axios, { AxiosError } from "axios";

export const api = axios.create({ baseURL: "/api/", withCredentials: true, headers: { Accept: "application/json" } });

export function setAccessToken(token: string | null) {
  if (token) api.defaults.headers.common.Authorization = `Bearer ${token}`;
  else delete api.defaults.headers.common.Authorization;
}

api.interceptors.response.use((response) => {
  if (response.status !== 204 && (typeof response.data !== "object" || response.data === null)) {
    return Promise.reject(new Error("Сервер вернул некорректный ответ"));
  }
  return response;
}, (cause: AxiosError<{ error?: string }>) => {
  if (axios.isCancel(cause)) return Promise.reject(cause);
  const message = cause.response?.data?.error;
  return Promise.reject(new Error(typeof message === "string" && message ? message : cause.response ? `Ошибка сервера (${cause.response.status})` : "Не удалось выполнить запрос"));
});
