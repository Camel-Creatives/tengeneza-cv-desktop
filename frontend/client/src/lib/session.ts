/** Short-lived browser session for the OAuth access token. Refresh tokens are never persisted. */
import type { User } from "./api";
export type Session = { token: string; refreshToken?: string; user: User };
const key = "tengenezacv_session";
export const session = {
  get: (): Session | null => {
    try {
      const raw = sessionStorage.getItem(key);
      return raw ? JSON.parse(raw) : null;
    } catch {
      return null;
    }
  },
  set: (v: Session) => sessionStorage.setItem(key, JSON.stringify({ ...v, refreshToken: undefined })),
  clear: () => sessionStorage.removeItem(key),
};
