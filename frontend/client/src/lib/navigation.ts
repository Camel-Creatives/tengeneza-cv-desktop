import { useLocation, useParams as useWouterParams } from "wouter";

export function usePathname() {
  return useLocation()[0].split("?")[0];
}

export function useRouter() {
  const [, navigate] = useLocation();
  return {
    push: (path: string) => navigate(path),
    replace: (path: string) => navigate(path, { replace: true }),
  };
}

export function useParams<T extends Record<string, string> = Record<string, string>>() {
  return useWouterParams<T>();
}
