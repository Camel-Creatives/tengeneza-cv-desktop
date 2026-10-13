"use client";

import { AppShell } from "@/components/AppShell";
import { Button } from "@/components/ui/button";
import { useLanguage } from "@/contexts/LanguageContext";
import { api, type CV, type Template, type User, type Wallet } from "@/lib/api";
import { getToken, useAuthGuard } from "@/lib/auth";
import {
  Copy,
  FilePlus2,
  FileText,
  LayoutTemplate,
  Loader2,
  Plus,
  Trash2,
  WalletCards,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

const blank = (u?: User) => ({
  personal: {
    full_name: u?.full_name || "",
    email: u?.email || "",
    phone: u?.phone || "",
    location: "",
    headline: "",
  },
  summary: "",
  experience: [],
  education: [],
  certifications: [],
  skills: { hard_skills: [], soft_skills: [], languages: [] },
  settings: { show_photo: false, show_referees: false },
});

type WorkspaceData = {
  user?: User;
  cvs: CV[];
  templates: Template[];
  wallet?: Wallet;
};

export function WorkspacePage() {
  const router = useRouter();
  useAuthGuard();
  const { t, locale } = useLanguage();
  const token = getToken();
  const [data, setData] = useState<WorkspaceData>({ cvs: [], templates: [] });
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [creating, setCreating] = useState(false);

  const loadWorkspace = useCallback(async () => {
    setLoading(true);
    setLoadError("");
    try {
      const [user, cvs, templates, wallet] = await Promise.all([
        api.me(token),
        api.listCVs(token),
        api.templates(token),
        api.wallet(token),
      ]);
      setData({ user, cvs: cvs.filter(cv => !cv.archived), templates, wallet });
    } catch (e) {
      const message =
        e instanceof Error ? e.message : "Unable to load your workspace.";
      setLoadError(message);
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    void loadWorkspace();
  }, [loadWorkspace]);

  async function create() {
    setCreating(true);
    try {
      const template_id =
        data.templates.find(x => !x.locked)?.id || "ats_classic";
      const cv = await api.createCV(token, {
        title: "Untitled CV",
        template_id,
        document_style: "resume",
        target_region: data.user?.target_region || "eac",
        language: locale,
        content: blank(data.user),
      });
      router.push(`/editor/${cv.id}`);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Unable to create a CV.");
    } finally {
      setCreating(false);
    }
  }

  async function duplicate(id: string) {
    try {
      const next = await api.duplicateCV(token, id);
      setData(d => ({ ...d, cvs: [next, ...d.cvs] }));
      toast.success("CV duplicated.");
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Could not duplicate this CV."
      );
    }
  }

  async function archive(id: string) {
    if (!window.confirm("Archive this CV? It will leave your workspace."))
      return;
    try {
      await api.archiveCV(token, id);
      setData(d => ({ ...d, cvs: d.cvs.filter(cv => cv.id !== id) }));
      toast.success("CV archived.");
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Could not archive this CV."
      );
    }
  }

  const firstName = data.user?.full_name?.trim().split(/\s+/)[0];
  const templatesAvailable = data.templates.filter(
    template => !template.locked
  ).length;
  const exportsRemaining = data.wallet?.daily_exports_remaining ?? -1;

  return (
    <AppShell>
      <div className="workspace-canvas px-4 py-7 sm:px-6 sm:py-9 xl:px-10">
        <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[#E5E7EB] pb-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-[#171717] sm:text-[28px]">
              {firstName
                ? t("Good to see you, {name}").replace("{name}", firstName)
                : t("Your CV workspace")}
            </h1>
            <p className="mt-1.5 text-sm text-[#737373]">
              {t("Create, improve, and keep your career documents in one place.")}
            </p>
          </div>
          <Button
            onClick={create}
            disabled={loading || creating}
            className="h-10 rounded-md bg-[#004AAD] px-4 font-semibold text-white hover:bg-[#003B8E]"
          >
            {creating ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Plus className="mr-2 h-4 w-4" />
            )}
            {creating ? t("Creating CV…") : t("create")}
          </Button>
        </div>

        <div className="mt-6 grid grid-cols-3 gap-2 sm:grid-cols-2 sm:gap-3 xl:grid-cols-3">
          <Metric
            icon={FileText}
            label={t("My CVs")}
            value={loading ? "—" : String(data.cvs.length)}
            note={t("Active documents")}
          />
          <Metric
            icon={LayoutTemplate}
            label={t("templates")}
            value={loading ? "—" : String(templatesAvailable)}
            note={t("Available to use")}
          />
          <Metric
            icon={WalletCards}
            label={t("Exports")}
            value={loading ? "—" : exportsRemaining < 0 ? t("Unlimited") : String(exportsRemaining)}
            note={exportsRemaining < 0 ? t("PDF, DOCX & CSV") : t("Available today")}
          />
        </div>

        <section className="mt-9" aria-labelledby="documents-heading">
          <div className="mb-4 flex items-end justify-between gap-4">
            <div>
              <h2
                id="documents-heading"
                className="text-lg font-semibold text-[#171717]"
              >
                {t("Your documents")}
              </h2>
              <p className="mt-1 text-sm text-[#737373]">
                {t("Pick up where you left off or start something new.")}
              </p>
            </div>
            <Link
              href="/templates"
              className="hidden items-center gap-1.5 text-sm font-medium text-[#004AAD] hover:underline sm:flex"
            >
              {t("browseTemplates")} <span aria-hidden="true">→</span>
            </Link>
          </div>

          {loading ? (
            <div
              className="grid gap-4 md:grid-cols-2 2xl:grid-cols-3"
              aria-label={t("My CVs")}
              aria-busy="true"
            >
              {[1, 2, 3].map(n => (
                <DocumentSkeleton key={n} />
              ))}
            </div>
          ) : loadError ? (
            <div className="rounded-lg border border-[#E5E7EB] bg-white px-6 py-12 text-center">
              <p className="font-semibold text-[#171717]">
                {t("Your documents could not be loaded")}
              </p>
              <p className="mx-auto mt-2 max-w-md text-sm text-[#737373]">
                {loadError}
              </p>
              <Button
                onClick={() => void loadWorkspace()}
                variant="outline"
                className="mt-5 border-[#D4D4D4] bg-white"
              >
                {t("Try again")}
              </Button>
            </div>
          ) : data.cvs.length ? (
            <div className="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
              {data.cvs.map(cv => {
                const template = data.templates.find(
                  item => item.id === cv.template_id
                );
                const personal = cv.content?.personal || {};
                const subtitle =
                  personal.headline ||
                  personal.full_name ||
                  t("Add your headline to help employers understand your experience.");
                return (
                  <article
                    key={cv.id}
                    className="group overflow-hidden rounded-lg border border-[#E5E7EB] bg-white transition duration-200 hover:-translate-y-0.5 hover:border-[#C9D1CC] hover:shadow-[0_8px_24px_rgba(0,0,0,.06)]"
                  >
                    <Link
                      href={`/editor/${cv.id}`}
                      className="block focus-visible:outline-offset-[-3px]"
                      aria-label={`Open ${cv.title}`}
                    >
                      <div className="document-card-preview" aria-hidden="true">
                        <div className="document-paper">
                          <span className="block h-1 w-10 rounded bg-[#004AAD]" />
                          <span className="mt-3 block h-2 w-28 max-w-full rounded-sm bg-[#262626]" />
                          <span className="mt-1.5 block h-1 w-20 rounded-sm bg-[#A3A3A3]" />
                          <span className="mt-4 block h-px w-full bg-[#E5E7EB]" />
                          <span className="mt-3 block h-1 w-full rounded-sm bg-[#D4D4D4]" />
                          <span className="mt-1.5 block h-1 w-4/5 rounded-sm bg-[#E5E7EB]" />
                          <span className="mt-3 block h-1 w-full rounded-sm bg-[#D4D4D4]" />
                          <span className="mt-1.5 block h-1 w-3/4 rounded-sm bg-[#E5E7EB]" />
                        </div>
                      </div>
                      <div className="px-4 pb-3 pt-4">
                        <div className="flex items-center justify-between gap-3">
                          <span className="inline-flex min-w-0 items-center gap-1.5 rounded-sm bg-[#F3F7FD] px-2 py-1 text-[11px] font-medium text-[#3F5146]">
                            <FileText
                              size={13}
                              className="shrink-0 text-[#004AAD]"
                            />
                            <span className="truncate">
                              {template?.name ||
                                cv.template_id.replaceAll("_", " ")}
                            </span>
                          </span>
                          <span className="shrink-0 text-[11px] text-[#737373]">
                            {cv.target_region?.toUpperCase() || "EAC"}
                          </span>
                        </div>
                        <h3 className="mt-3 truncate text-base font-semibold text-[#171717] group-hover:text-[#004AAD]">
                          {cv.title}
                        </h3>
                        <p className="mt-1 min-h-10 line-clamp-2 text-sm leading-5 text-[#737373]">
                          {subtitle}
                        </p>
                      </div>
                    </Link>
                    <div className="flex items-center justify-between border-t border-[#F0F0F0] px-3 py-2">
                      <Link
                        href={`/editor/${cv.id}`}
                        className="rounded px-2 py-1.5 text-xs font-semibold text-[#004AAD] hover:bg-[#F3F7FD]"
                      >
                        {t("Open CV")} <span aria-hidden="true">→</span>
                      </Link>
                      <div className="flex items-center">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => void duplicate(cv.id)}
                          title={t("Duplicate CV")}
                          aria-label={`Duplicate ${cv.title}`}
                          className="h-8 w-8 text-[#737373] hover:text-[#171717]"
                        >
                          <Copy size={15} />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => void archive(cv.id)}
                          title={t("Archive CV")}
                          aria-label={`Archive ${cv.title}`}
                          className="h-8 w-8 text-[#737373] hover:bg-red-50 hover:text-[#B42318]"
                        >
                          <Trash2 size={15} />
                        </Button>
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>
          ) : (
            <div className="rounded-lg border border-dashed border-[#C9D1CC] bg-white px-6 py-14 text-center sm:py-16">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-[#EAF2FF] text-[#004AAD]">
                <FilePlus2 size={22} />
              </div>
              <h3 className="mt-4 text-lg font-semibold text-[#171717]">
                {t("Your first CV starts here")}
              </h3>
              <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-[#737373]">
                {t("Choose a professional template, add your experience, then review and export your CV when it is ready.")}
              </p>
              <Button
                onClick={create}
                disabled={creating}
                className="mt-5 h-10 bg-[#004AAD] px-4 font-semibold text-white hover:bg-[#003B8E]"
              >
                {creating ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Plus className="mr-2 h-4 w-4" />
                )}
                {t("Create your first CV")}
              </Button>
              <Link
                href="/templates"
                className="ml-3 inline-flex h-10 items-center rounded-md px-3 text-sm font-medium text-[#525252] hover:bg-[#F5F5F5]"
              >
                {t("Explore templates")}
              </Link>
            </div>
          )}
        </section>
        <div className="mt-6 flex justify-center sm:hidden">
          <Link
            href="/templates"
            className="text-sm font-medium text-[#004AAD]"
          >
            {t("Browse all templates")} →
          </Link>
        </div>
      </div>
    </AppShell>
  );
}

function Metric({
  icon: Icon,
  label,
  value,
  note,
}: {
  icon: typeof FileText;
  label: string;
  value: string;
  note: string;
}) {
  return (
    <article className="flex min-h-[74px] items-center gap-2 rounded-lg border border-[#E5E7EB] bg-white px-2 py-2 sm:min-h-[88px] sm:gap-3 sm:px-4 sm:py-3.5">
      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-[#F3F7FD] text-[#004AAD] sm:h-9 sm:w-9">
        <Icon size={15} className="sm:h-[17px] sm:w-[17px]" />
      </span>
      <div className="min-w-0">
        <div className="flex items-baseline gap-2">
          <span className="text-lg font-semibold tabular-nums text-[#171717] sm:text-xl">
            {value}
          </span>
          <span className="truncate text-[10px] font-medium text-[#525252] sm:text-xs">
            {label}
          </span>
        </div>
        <p className="mt-0.5 truncate text-[9px] text-[#737373] sm:text-xs">{note}</p>
      </div>
    </article>
  );
}

function DocumentSkeleton() {
  return (
    <div
      className="overflow-hidden rounded-lg border border-[#E5E7EB] bg-white"
      aria-hidden="true"
    >
      <div className="h-[150px] animate-pulse bg-[#F3F7FD]" />
      <div className="space-y-3 p-4">
        <div className="h-4 w-28 animate-pulse rounded bg-[#F0F0F0]" />
        <div className="h-5 w-2/3 animate-pulse rounded bg-[#F0F0F0]" />
        <div className="h-4 w-full animate-pulse rounded bg-[#F0F0F0]" />
        <div className="h-8 w-full animate-pulse rounded bg-[#F8F8F8]" />
      </div>
    </div>
  );
}
