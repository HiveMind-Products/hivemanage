import { LoginSchema } from "@/typings/auth";
import { fetchApi } from "@/utils/http-util";
import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "react-router";

export function useLogin() {
  const navigate = useNavigate();
  return useMutation({
    mutationFn: (data: LoginSchema) =>
      fetchApi("/api/dash/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
        credentials: "include",
      }),
    onSuccess: () => {
      navigate("/app");
    },
  });
}
