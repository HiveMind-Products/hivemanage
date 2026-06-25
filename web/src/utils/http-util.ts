export const IS_DEV = import.meta.env.DEV;

type ApiResponse<T> = { status: string; data: T };
export class ApiError extends Error { constructor(error: string, errorMessage: string) { super(errorMessage); this.name = error; } }
export function getCookie(name: string): string | undefined { return document.cookie.split(";").map((value) => value.trim()).find((value) => value.startsWith(`${name}=`))?.split("=").slice(1).join("="); }
function isMutatingMethod(method: string | undefined) { return ["POST", "PUT", "PATCH", "DELETE"].includes((method ?? "GET").toUpperCase()); }
export async function fetchApi<T = unknown>(input: string | URL, init?: RequestInit): Promise<T | undefined> {
  const headers = new Headers(init?.headers);
  const isFormData = init?.body instanceof FormData;
  if (!isFormData && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const url = input.toString();
  if (url.startsWith("/api/dash") && isMutatingMethod(init?.method)) {
    const csrfToken = getCookie("fmlite_csrf");
    if (csrfToken) headers.set("X-CSRF-Token", decodeURIComponent(csrfToken));
  }
  const res = await fetch(input, { ...init, credentials: "include", headers });
  if (!res.ok) {
    const errorResponse = (await res.json().catch(() => undefined)) as { error?: string; message?: string; data?: { message?: string } } | undefined;
    const message = errorResponse?.message ?? errorResponse?.data?.message ?? res.statusText;
    throw new ApiError(errorResponse?.error ?? "ApiError", message);
  }
  if (res.status === 204) return undefined;
  const response = (await res.json()) as ApiResponse<T>;
  return response.data;
}
