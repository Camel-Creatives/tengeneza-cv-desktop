import type { User } from "./api";

/** Desktop mode has one local workspace and never asks the user to sign in. */
declare global {
  interface Window {
    camel?: {
      identify: (accountId: string) => void;
      clearIdentity: () => void;
    };
  }
}

export async function createCamelAuthorizePath(_returnTo = "/workspace") {
  return undefined;
}

export async function startCamelLogin(_returnTo = "/workspace") {
  return undefined;
}

export async function startCamelSocialLogin(_provider: string, _returnTo = "/workspace") {
  return undefined;
}

// Kept as a compatibility hook for pages inherited from the web client.
export function useAuthGuard() {}

// The local API intentionally receives no bearer token. Its loopback-only
// desktop mode selects the current device's workspace automatically.
export function getToken(): string {
  return "";
}

export function getUser(): User | null {
  return null;
}

export function signOut() {
  if (typeof window !== "undefined") window.location.assign("/workspace");
}
