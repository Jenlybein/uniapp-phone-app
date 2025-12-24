import { defineStore } from "pinia";
import axios from "axios";

interface User {
  id: number;
  username: string;
  email: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthState => ({
    user: null,
    token: localStorage.getItem("token"),
    isAuthenticated: !!localStorage.getItem("token"),
    loading: false,
    error: null,
  }),

  actions: {
    async login(username: string, password: string) {
      this.loading = true;
      this.error = null;

      try {
        const response = await axios.post(
          "http://localhost:8080/api/auth/login",
          {
            username,
            password,
          }
        );

        // 检查响应是否成功
        if (response.status === 200 && response.data.code === 200) {
          const { token, user } = response.data.data;

          this.token = token;
          this.user = user;
          this.isAuthenticated = true;

          localStorage.setItem("token", token);
          localStorage.setItem("user", JSON.stringify(user));

          return true;
        } else {
          this.error = response.data?.message || "登录失败，请检查用户名和密码";
          return false;
        }
      } catch (error: any) {
        console.error("登录错误:", error);
        this.error =
          error.response?.data?.message || "登录失败，请检查用户名和密码";
        return false;
      } finally {
        this.loading = false;
      }
    },

    async register(username: string, email: string, password: string) {
      this.loading = true;
      this.error = null;

      try {
        const response = await axios.post(
          "http://localhost:8080/api/auth/register",
          {
            username,
            email,
            password,
          }
        );

        // 检查响应是否成功
        if (response.status === 200 && response.data.code === 200) {
          const { token } = response.data.data;

          this.token = token;
          this.isAuthenticated = true;

          localStorage.setItem("token", token);

          return true;
        } else {
          this.error = response.data?.message || "注册失败，请稍后重试";
          return false;
        }
      } catch (error: any) {
        console.error("注册错误:", error);
        this.error = error.response?.data?.message || "注册失败，请稍后重试";
        return false;
      } finally {
        this.loading = false;
      }
    },

    logout() {
      this.user = null;
      this.token = null;
      this.isAuthenticated = false;

      localStorage.removeItem("token");
      localStorage.removeItem("user");
    },

    async refreshToken() {
      const currentToken = this.token;

      if (!currentToken) {
        this.logout();
        return;
      }

      try {
        const response = await axios.post(
          "http://localhost:8080/api/auth/refresh",
          {
            token: currentToken,
          }
        );

        const { token } = response.data.data;

          this.token = token;
          localStorage.setItem("token", token);
      } catch (error) {
        this.logout();
      }
    },
  },
});
