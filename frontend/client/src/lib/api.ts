/** Nairobi Blue Ledger: typed API boundary for a document-first workspace. */
const getEnvApiUrl = () => {
  if (typeof process !== "undefined" && process.env) {
    return process.env.NEXT_PUBLIC_TENGENEZACV_API_BASE_URL || process.env.VITE_TENGENEZACV_API_BASE_URL;
  }
  return undefined;
};

const previewApiBase =
  typeof window !== "undefined" && /^3000-/.test(window.location.host)
    ? `${window.location.protocol}//${window.location.host.replace(/^3000-/, "8080-")}`
    : typeof window !== "undefined" && (window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1")
    ? "http://localhost:8080"
    : "";

let desktopApiBaseUrl = "";
export function setDesktopAPIBaseURL(url: string) {
  if (!/^http:\/\/127\.0\.0\.1:\d+$/.test(url)) {
    throw new Error("The desktop API must use a private loopback address.");
  }
  desktopApiBaseUrl = url.replace(/\/$/, "");
}
function currentApiBaseURL() {
  return desktopApiBaseUrl || API_BASE_URL;
}

export const API_BASE_URL = (
  getEnvApiUrl() ||
  previewApiBase ||
  "https://api.tengenezacv.online"
).replace(/\/$/, "");
export function assetUrl(url?: string): string {
  if (!url) return "";
  if (/^(data:|blob:)/.test(url)) return url;

  // Handle document exports (e.g. file:///tmp/tengenezacv-exports/Brian_Lema_CV.pdf)
  if (url.startsWith("file://") && (url.endsWith(".pdf") || url.endsWith(".docx") || url.endsWith(".csv"))) {
    const filename = url.split("/").pop() || "";
  return `${currentApiBaseURL()}/exports/${filename}`;
  }

  // Clean up legacy file:// or local path references for images
  if (url.startsWith("file://") || url.includes("/profile-assets/")) {
    const filename = url.split("/").pop() || "";
    return `${currentApiBaseURL()}/uploads/profiles/${filename}`;
  }

  if (/^https?:\/\//.test(url)) return url;

  if (url.startsWith("/examples/")) {
    return url;
  }

  return `${currentApiBaseURL()}${url.startsWith("/") ? "" : "/"}${url}`;
}
export type User = {
  id: string;
  full_name: string;
  email?: string;
  phone?: string;
  profile_photo_url?: string;
  preferred_language?: string;
  target_region?: string;
  plan?: string;
  role?: string;
};
export type Wallet = {
  export_credits: number;
  daily_export_limit: number;
  daily_exports_remaining: number;
  ai_token_balance: number;
  voice_seconds_balance: number;
};
export type Template = {
  id: string;
  name: string;
  family: string;
  supports_photo: boolean;
  locked: boolean;
  accent_color_hex: string;
};
export type ExampleSummary = {
  slug: string;
  title: string;
  template_id: string;
  full_name?: string;
  headline?: string;
};
export type CVContent = Record<string, any>;
export type CV = {
  id: string;
  title: string;
  template_id: string;
  document_style: string;
  target_region: string;
  language: string;
  content: CVContent;
  archived?: boolean;
};
export type AtsReport = {
  score: number;
  breakdown: {
    check: string;
    severity: "info" | "warning" | "error";
    message: string;
  }[];
  coach_note?: string;
};
export type Plan = {
  code: string;
  name: string;
  price_tzs?: number;
  price?: number;
  export_credits?: number;
  ai_tokens?: number;
  voice_seconds?: number;
  subscription?: boolean;
  description?: string;
  features?: string[];
};
type Envelope<T> = {
  success: boolean;
  data: T;
  error?: { code: string; message: string };
  meta?: { request_id: string };
};
export class TengenezaApiError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string,
    public requestId?: string
  ) {
    super(message);
  }
}
function normalizeCV(raw: any): CV {
  return {
    id: raw.id ?? raw.ID,
    title: raw.title ?? raw.Title,
    template_id: raw.template_id ?? raw.TemplateID,
    document_style: raw.document_style ?? raw.DocumentStyle,
    target_region: raw.target_region ?? raw.TargetRegion,
    language: raw.language ?? raw.Language,
    content: raw.content ?? raw.Content ?? {},
    archived: raw.archived ?? raw.Archived,
  };
}
function normalizePlan(raw: any): Plan {
  return {
    code: raw.code ?? raw.Code,
    name: raw.name ?? raw.label ?? raw.Label,
    price_tzs: raw.price_tzs ?? raw.AmountTZS,
    export_credits: raw.export_credits ?? raw.ExportCredits,
    ai_tokens: raw.ai_tokens ?? raw.AITokenGrant,
    voice_seconds: raw.voice_seconds ?? raw.VoiceSeconds,
    subscription: raw.subscription ?? raw.Subscription ?? false,
    description: raw.description ?? raw.Label,
    features: raw.features,
  };
}
function normalizeAts(raw: any): AtsReport {
  return {
    score: raw.score ?? raw.Score ?? 0,
    breakdown: (raw.breakdown ?? raw.Breakdown ?? []).map((item: any) => ({
      check: item.check ?? item.Check,
      severity: item.severity ?? item.Severity,
      message: item.message ?? item.Message,
    })),
    coach_note: raw.coach_note ?? raw.CoachNote,
  };
}
function normalizeVersion(raw: any) {
  return {
    ...raw,
    id: raw.id ?? raw.ID,
    change_summary: raw.change_summary ?? raw.ChangeSummary,
  };
}
async function request<T>(
  path: string,
  init: RequestInit = {},
  token?: string,
  idem?: string
): Promise<T> {
  const headers = new Headers(init.headers);
  if (!(init.body instanceof FormData))
    headers.set("Content-Type", "application/json");
  headers.set("X-Request-ID", crypto.randomUUID());
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (idem) headers.set("Idempotency-Key", idem);
  let response: Response;
  try {
    response = await fetch(`${currentApiBaseURL()}${path}`, { ...init, headers });
  } catch {
    throw new TengenezaApiError(
      "We could not reach the Tengeneza CV service. Check the configured API URL."
    );
  }
  const raw = await response.text();
  let payload: Envelope<T>;
  try {
    payload = JSON.parse(raw);
  } catch {
    throw new TengenezaApiError(
      "The configured URL did not return the Tengeneza CV API response format.",
      response.status
    );
  }
  if (!response.ok || !payload.success)
    throw new TengenezaApiError(
      payload.error?.message || "The request could not be completed.",
      response.status,
      payload.error?.code,
      payload.meta?.request_id
    );
  return payload.data;
}
export const api = {
  signup: (x: {
    full_name: string;
    email?: string;
    phone?: string;
    password?: string;
  }) =>
    request<{
      verification_id: string;
      destination: string;
      development_otp?: string;
    }>("/auth/signup", {
      method: "POST",
      body: JSON.stringify(x),
    }),
  verifyOtp: (x: { verification_id: string; code: string }) =>
    request<{ access_token: string; refresh_token: string; user: User }>(
      "/auth/otp/verify",
      { method: "POST", body: JSON.stringify(x) }
    ),
  resendOtp: (x: { verification_id: string }) =>
    request<{
      verification_id: string;
      destination: string;
      expires_in_seconds: number;
      development_otp?: string;
    }>("/auth/otp/resend", {
      method: "POST",
      body: JSON.stringify(x),
    }),
  login: (x: { email: string; password: string }) =>
    request<{ access_token: string; refresh_token: string; user: User }>(
      "/auth/login/password",
      { method: "POST", body: JSON.stringify(x) }
    ),
  adminLogin: (x: { email: string; password: string }) =>
    request<{ access_token: string; refresh_token: string; user: User }>(
      "/auth/admin/login",
      { method: "POST", body: JSON.stringify(x) }
    ),
  me: (t: string) => request<User>("/me", {}, t),
  updateMe: (t: string, x: Partial<User>) =>
    request<User>("/me", { method: "PATCH", body: JSON.stringify(x) }, t),
  uploadPhoto: (t: string, photo: File) => {
    const f = new FormData();
    f.append("photo", photo);
    return request<{ profile_photo_url: string }>(
      "/me/profile-photo",
      { method: "POST", body: f },
      t
    );
  },
  wallet: (t: string) => request<Wallet>("/me/wallet", {}, t),
  templates: (t: string) => request<Template[]>("/templates", {}, t),
  listCVs: async (t: string) =>
    (await request<any[]>("/cvs", {}, t)).map(normalizeCV),
  getCV: async (t: string, id: string) =>
    normalizeCV(await request<any>(`/cvs/${id}`, {}, t)),
  createCV: async (t: string, x: Partial<CV>) =>
    normalizeCV(
      await request<any>("/cvs", { method: "POST", body: JSON.stringify(x) }, t)
    ),
  patchCV: async (
    t: string,
    id: string,
    content: CVContent,
    change_summary: string
  ) =>
    normalizeCV(
      await request<any>(
        `/cvs/${id}`,
        { method: "PATCH", body: JSON.stringify({ content, change_summary }) },
        t
      )
    ),
  archiveCV: (t: string, id: string) =>
    request<{ archived: boolean }>(`/cvs/${id}`, { method: "DELETE" }, t),
  duplicateCV: async (t: string, id: string) =>
    normalizeCV(
      await request<any>(`/cvs/${id}/duplicate`, { method: "POST" }, t)
    ),
  versions: async (t: string, id: string) =>
    (await request<any[]>(`/cvs/${id}/versions`, {}, t)).map(normalizeVersion),
  restoreVersion: async (t: string, id: string, v: string) =>
    normalizeCV(
      await request<any>(
        `/cvs/${id}/versions/${v}/restore`,
        { method: "POST" },
        t
      )
    ),
  applyProfilePhoto: async (t: string, id: string) =>
    normalizeCV(
      await request<any>(`/cvs/${id}/use-profile-photo`, { method: "POST" }, t)
    ),
  ats: async (t: string, id: string) =>
    normalizeAts(
      await request<any>(`/cvs/${id}/agent/ats-score`, { method: "POST" }, t)
    ),
  atsCheck: async (t: string, id: string) =>
    normalizeAts(
      await request<any>(`/cvs/${id}/agent/ats-score`, { method: "POST" }, t)
    ),
  draft: (t: string, id: string, section: string, instruction: string) =>
    request<{ draft: string }>(
      `/cvs/${id}/agent/autofill-section`,
      { method: "POST", body: JSON.stringify({ section, instruction }) },
      t
    ),
  aiDraft: (
    t: string,
    x: { cv_id: string; section: string; instruction: string }
  ) =>
    request<{ draft: string }>(
      `/cvs/${x.cv_id}/agent/autofill-section`,
      { method: "POST", body: JSON.stringify({ section: x.section, instruction: x.instruction }) },
      t
    ),
  exportCV: (t: string, id: string, format: "pdf" | "docx" | "csv", language?: "en" | "sw") =>
    request<{
      file_url: string;
      download_url?: string;
      filename: string;
      bytes?: number;
      credit_consumed?: boolean;
      daily_export_limit: number;
      daily_exports_remaining: number;
    }>(`/cvs/${id}/export`, { method: "POST", body: JSON.stringify({ format, language }) }, t),
  export: (t: string, id: string, format: "pdf" | "docx" | "csv") =>
    request<{
      file_url: string;
      download_url?: string;
      filename: string;
      bytes?: number;
      credit_consumed?: boolean;
      daily_export_limit: number;
      daily_exports_remaining: number;
    }>(`/cvs/${id}/export`, { method: "POST", body: JSON.stringify({ format }) }, t),
  emailExport: (t: string, id: string, email: string, filename: string) =>
    request<{ sent: boolean }>(`/cvs/${id}/export/email`, { method: "POST", body: JSON.stringify({ email, filename }) }, t),
  plans: async () =>
    (await request<any[]>("/billing/plans")).map(normalizePlan),
  templatesPublic: () => request<Template[]>("/templates"),
  examples: () => request<ExampleSummary[]>("/examples"),
  example: (slug: string) =>
    request<{ title: string; template_id: string; content: CVContent }>(
      `/examples/${slug}`
    ),
};

export async function downloadFile(url: string, filename: string, token?: string): Promise<void> {
  try {
    // The API filename is authoritative; file_url may be an object key or a
    // legacy local path and download_url may include query parameters.
    const exportFilename = filename.split(/[/?#]/).filter(Boolean).pop() || filename;
    const target = token && /\.(pdf|docx|csv)$/i.test(exportFilename)
      ? `${currentApiBaseURL()}/exports/${encodeURIComponent(exportFilename)}`
      : assetUrl(url);
    const response = await fetch(target, { headers: token ? { Authorization: `Bearer ${token}` } : undefined });
    if (!response.ok) {
      let message = `Download failed (${response.status}).`;
      try {
        const payload = await response.json();
        message = payload?.error?.message || payload?.message || message;
      } catch {}
      throw new Error(message);
    }
    const blob = await response.blob();
    const href = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = href;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    setTimeout(() => URL.revokeObjectURL(href), 4000);
  } catch (error) {
    if (token) {
      throw new Error(error instanceof Error ? error.message : "The export download failed.");
    }
    window.open(assetUrl(url), "_blank", "noopener,noreferrer");
  }
}
