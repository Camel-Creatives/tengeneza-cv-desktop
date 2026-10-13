'use client';

/** Nairobi Blue Ledger: real API-backed editing, ATS review, version history, AI drafting, and export with single-line top preview toolbar. */
import { AppShell } from "@/components/AppShell";
import { CvPreview } from "@/components/CvPreview";
import { useLanguage } from "@/contexts/LanguageContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { api, assetUrl, AtsReport, CV, CVContent, downloadFile, Template } from "@/lib/api";
import { getToken, useAuthGuard } from "@/lib/auth";
import {
  AlertTriangle,
  ArrowLeft,
  Camera,
  Check,
  ChevronRight,
  Download,
  Eye,
  ExternalLink,
  FileClock,
  Loader2,
  Mail,
  MoreHorizontal,
  Palette,
  Plus,
  Save,
  Sparkles,
  Trash2,
  Upload,
  User,
  Wand2,
} from "lucide-react";
import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";

const list = (x: any) => (Array.isArray(x) ? x : []);
const splitSkills = (x: string) =>
  x.split(",").map(s => s.trim()).filter(Boolean);

const PRESET_COLORS = [
  "#004AAD", // Tengeneza cobalt
  "#003B8E", // Deep cobalt
  "#2563EB", // Bright blue
  "#5B8DEF", // Clear blue
  "#759FEA", // Soft blue
  "#9DBCF2", // Pale blue
];

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs font-extrabold text-[#525252]">{label}</Label>
      {children}
    </div>
  );
}

function Section({
  label,
  children,
  className,
}: {
  label: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={className}>
      <div className="mb-4 flex items-center gap-3">
        <i className="h-px w-5 bg-[#004AAD]" />
        <h2 className="text-sm font-extrabold uppercase tracking-[.13em]">
          {label}
        </h2>
      </div>
      {children}
    </section>
  );
}

export function EditorPage() {
  const router = useRouter();
  const { t, locale } = useLanguage();
  const params = useParams();
  const id = (params?.id as string) || "";
  useAuthGuard();
  const token = getToken();
  const [cv, setCv] = useState<CV | null>(null),
    [templates, setTemplates] = useState<Template[]>([]),
    [report, setReport] = useState<AtsReport | null>(null),
    [history, setHistory] = useState<any[] | null>(null),
    [instruction, setInstruction] = useState(""),
    [draft, setDraft] = useState(""),
    [saving, setSaving] = useState(false),
    [exportingFormat, setExportingFormat] = useState<"pdf" | "docx" | "csv" | null>(null),
    [switchingTemplate, setSwitchingTemplate] = useState(false),
    [applyingPhoto, setApplyingPhoto] = useState(false),
    [uploadingPhoto, setUploadingPhoto] = useState(false),
    [hardSkillsDraft, setHardSkillsDraft] = useState<string | null>(null),
    [softSkillsDraft, setSoftSkillsDraft] = useState<string | null>(null),
    [languagesDraft, setLanguagesDraft] = useState<string | null>(null);
  const [mobileEditorView, setMobileEditorView] = useState<"templates" | "editor" | "export">("templates");
  const [selectedTemplateId, setSelectedTemplateId] = useState("");

  // Top Preview Toolbar Popovers state
  const [customAccentColor, setCustomAccentColor] = useState<string | null>(null);
  const [showColorPicker, setShowColorPicker] = useState(false);
  const [showAtsPopover, setShowAtsPopover] = useState(false);
  const [showHistoryPopover, setShowHistoryPopover] = useState(false);
  const [showMoreActions, setShowMoreActions] = useState(false);
  const [deliveryEmail, setDeliveryEmail] = useState("");
  const [lastExportFilename, setLastExportFilename] = useState("");
  const [emailSending, setEmailSending] = useState(false);
  const [showEmailPanel, setShowEmailPanel] = useState(false);

  const colorPickerRef = useRef<HTMLDivElement>(null);
  const atsRef = useRef<HTMLDivElement>(null);
  const historyRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    try {
      const savedView = sessionStorage.getItem("tengenezacv_mobile_editor_view");
      if (savedView === "editor" || savedView === "export") {
        setMobileEditorView(savedView);
      }
    } catch {}
    if (!id) return router.push("/workspace");
     Promise.all([api.getCV(token, id), api.templates(token)])
      .then(([x, y]) => {
        let recovered = x;
        try {
          const raw = localStorage.getItem(`tengenezacv_cv_draft_${id}`);
          if (raw) {
            const saved = JSON.parse(raw) as { cv?: CV };
            if (saved.cv?.id === x.id) recovered = saved.cv;
          }
        } catch {}
        setCv(recovered);
        setTemplates(y);
        setSelectedTemplateId(recovered.template_id);
      })
      .catch(e => {
        toast.error(e instanceof Error ? e.message : "Unable to open CV.");
        router.push("/workspace");
      });
  }, [id, router]);

  // Keep an encrypted-session-independent local draft so a refresh or a
  // mobile browser tab discard cannot erase work that has not been saved yet.
  useEffect(() => {
    if (!cv || !id) return;
    try {
      localStorage.setItem(`tengenezacv_cv_draft_${id}`, JSON.stringify({ cv, savedAt: new Date().toISOString() }));
      sessionStorage.setItem("tengenezacv_mobile_editor_view", mobileEditorView);
    } catch {}
  }, [cv, id, mobileEditorView]);

  // Click outside listener for popovers
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (colorPickerRef.current && !colorPickerRef.current.contains(e.target as Node)) {
        setShowColorPicker(false);
      }
      if (atsRef.current && !atsRef.current.contains(e.target as Node)) {
        setShowAtsPopover(false);
      }
      if (historyRef.current && !historyRef.current.contains(e.target as Node)) {
        setShowHistoryPopover(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  if (!cv)
    return (
      <AppShell>
        <div className="flex min-h-[60vh] items-center justify-center">
          <Loader2 className="animate-spin text-[#004AAD]" />
        </div>
      </AppShell>
    );

  const content = cv.content || {},
    personal = content.personal || {},
    exp = list(content.experience),
    edu = list(content.education),
    certs = list(content.certifications),
    referees = list(content.referees),
    skills = content.skills || {};

  const activeTemplate = templates.find(t => t.id === cv.template_id);
  const activeColor = customAccentColor || activeTemplate?.accent_color_hex || "#004AAD";

  const change = (next: CVContent) => setCv({ ...cv, content: next });
  const personalChange = (key: string, value: string) =>
    change({ ...content, personal: { ...personal, [key]: value } });
  const experienceChange = (i: number, key: string, value: any) =>
    change({
      ...content,
      experience: exp.map((x: any, n: number) =>
        n === i ? { ...x, [key]: value } : x
      ),
    });
  const educationChange = (i: number, key: string, value: any) =>
    change({
      ...content,
      education: edu.map((x: any, n: number) =>
        n === i ? { ...x, [key]: value } : x
      ),
    });
  const certificationChange = (i: number, key: string, value: any) =>
    change({
      ...content,
      certifications: certs.map((x: any, n: number) =>
        n === i ? { ...x, [key]: value } : x
      ),
    });
  const refereeChange = (i: number, key: string, value: any) =>
    change({
      ...content,
      referees: referees.map((x: any, n: number) =>
        n === i ? { ...x, [key]: value } : x
      ),
    });

  const getBulletsText = (x: any): string => {
    const b = x.bullets || x.achievements || x.highlights || x.responsibilities;
    if (Array.isArray(b)) {
      return b.map(item => (typeof item === "string" ? item : String(item?.item || item?.text || item))).join("\n");
    }
    if (typeof b === "string") return b;
    if (typeof x.description === "string") return x.description;
    return "";
  };

  const handleBulletsChange = (i: number, text: string) => {
    const bulletLines = text.split("\n");
    change({
      ...content,
      experience: exp.map((item: any, idx: number) => {
        if (idx !== i) return item;
        return {
          ...item,
          bullets: bulletLines,
          achievements: bulletLines,
          highlights: bulletLines,
        };
      }),
    });
  };

  const roleOf = (x: any) => x.title ?? x.role ?? "";
  const employerOf = (x: any) => x.employer ?? x.company ?? "";
  const credentialOf = (x: any) => x.credential ?? x.qualification ?? "";

  async function save(summary = "Updated CV") {
    if (!cv) return false;
    let contentToPersist = { ...cv.content };
    const currentSkills = contentToPersist.skills || {};
    let nextSkills = { ...currentSkills };
    if (hardSkillsDraft !== null) {
      nextSkills.hard_skills = splitSkills(hardSkillsDraft);
    }
    if (softSkillsDraft !== null) {
      nextSkills.soft_skills = splitSkills(softSkillsDraft);
    }
    if (languagesDraft !== null) {
      nextSkills.languages = splitSkills(languagesDraft);
    }
    contentToPersist.skills = nextSkills;

    setSaving(true);
    try {
      setCv(
        await api.patchCV(token, cv.id, contentToPersist, summary)
      );
      setHardSkillsDraft(null);
      setSoftSkillsDraft(null);
      setLanguagesDraft(null);
      toast.success(t("Changes saved."));
      return true;
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Unable to save your CV."
      );
      return false;
    } finally {
      setSaving(false);
    }
  }

  async function runScore() {
    if (!cv) return;
    try {
      setReport(await api.atsCheck(token, cv.id));
      toast.success(t("ATS score updated."));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "ATS check failed.");
    }
  }

  async function exportCV(format: "pdf" | "docx" | "csv") {
    if (!cv || exportingFormat) return;
    setExportingFormat(format);
    const ok = await save(`Export prep (${format.toUpperCase()})`);
    if (!ok) {
      setExportingFormat(null);
      return;
    }
    try {
      const res = await api.exportCV(token, cv.id, format, locale);
      setLastExportFilename(res.filename);
      const downloadTarget = res.download_url || res.file_url;
      if (downloadTarget) {
        await downloadFile(downloadTarget, res.filename || `${cv.title || "CV"}.${format}`, token);
        toast.success(`Exported ${res.filename || format.toUpperCase()}`);
      } else {
        toast.error("Export file link not available.");
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Export failed.");
    } finally {
      setExportingFormat(null);
    }
  }

  async function emailLatestExport() {
    if (!cv || !deliveryEmail.includes("@") || !lastExportFilename || emailSending) return;
    setEmailSending(true);
    try { await api.emailExport(token, cv.id, deliveryEmail.trim(), lastExportFilename); toast.success("Your CV was sent by email."); }
    catch (e) { toast.error(e instanceof Error ? e.message : "The CV email could not be sent."); }
    finally { setEmailSending(false); }
  }

  async function draftSection(section: string) {
    setDraft("Drafting with AI...");
    try {
      const r = await api.aiDraft(token, {
        cv_id: cv?.id || id,
        instruction,
        section,
      });
      setDraft(r.draft);
      toast.success("Draft ready for review.");
    } catch (e) {
      setDraft("");
      toast.error(e instanceof Error ? e.message : "AI drafting failed.");
    }
  }

  async function uploadPhoto(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0];
    e.target.value = "";
    if (!cv || !f) return;
    if (!["image/jpeg", "image/png", "image/webp"].includes(f.type)) {
      toast.error("Choose a JPEG, PNG, or WebP image.");
      return;
    }
    if (f.size > 5 * 1024 * 1024) {
      toast.error("Choose an image smaller than 5 MB.");
      return;
    }
    setUploadingPhoto(true);
    try {
      const res = await api.uploadPhoto(token, f);
      const newPhotoUrl = res.profile_photo_url;
      change({
        ...content,
        personal: {
          ...personal,
          photo_url: newPhotoUrl,
        },
      });
      toast.success("Photo uploaded and applied to this CV.");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Photo upload failed.");
    } finally {
      setUploadingPhoto(false);
    }
  }

  async function loadHistory() {
    if (!cv) return;
    try {
      setHistory(await api.versions(token, cv.id));
    } catch {
      setHistory([]);
    }
  }

  async function restore(v: any) {
    if (
      !cv ||
      !window.confirm(
        "Restore this version? It will replace current CV content."
      )
    )
      return;
    try {
      setCv(await api.restoreVersion(token, cv.id, v.id));
      setShowHistoryPopover(false);
      toast.success("Version restored.");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Could not restore.");
    }
  }

  async function switchTemplate(template_id: string) {
    if (!cv || template_id === cv.template_id) return;
    const selected = templates.find(template => template.id === template_id);
    setSwitchingTemplate(true);
    try {
      const edition = await api.createCV(token, {
        title: `${cv.title} — ${selected?.name || "Template edition"}`,
        template_id,
        document_style: cv.document_style,
        target_region: cv.target_region,
        language: cv.language,
        content: cv.content,
      });
      toast.success(
        "New template edition created. Your original CV remains unchanged."
      );
      try { sessionStorage.setItem("tengenezacv_mobile_editor_view", "editor"); } catch {}
      router.push(`/editor/${edition.id}`);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Could not switch template."
      );
    } finally {
      setSwitchingTemplate(false);
    }
  }

  function selectTemplate(template_id: string) {
    setSelectedTemplateId(template_id);
    if (typeof window !== "undefined" && window.innerWidth < 768) {
      void switchTemplate(template_id);
    }
  }

  return (
    <AppShell>
      <div className={`mobile-template-picker ${mobileEditorView !== "templates" ? "mobile-template-picker-hidden" : ""}`}>
        <Link href="/workspace" className="mb-6 inline-flex items-center gap-2 text-sm font-bold text-[#525252]"><ArrowLeft size={17} /> {t("workspace")}</Link>
        <p className="ledger-label"><span /> {t("Template Design")}</p>
        <h1 className="mt-3 text-2xl font-extrabold text-[#171717]">{t("Choose your CV design")}</h1>
        <p className="mt-2 text-sm leading-6 text-[#525252]">{t("Select a design, review your CV, then continue to edit.")}</p>
        <div className="template-choice-layout mt-6">
          <div className="template-choice-list grid grid-cols-2 gap-3">
            {templates.map(t => <button key={t.id} type="button" onClick={() => selectTemplate(t.id)} disabled={switchingTemplate} className={`min-h-20 rounded-lg border p-3 text-left transition ${selectedTemplateId === t.id ? "border-[#004AAD] bg-[#EAF2FF]" : "border-[#E5E7EB] bg-white active:bg-[#EAF2FF]"}`}><b className="block text-sm text-[#171717]">{t.name}</b><span className="mt-1 block text-[11px] capitalize text-[#737373]">{t.family.replace("_", " ")}</span></button>)}
          </div>
          <aside className="template-choice-preview rounded-xl border border-[#E5E7EB] bg-white p-4 shadow-xs">
            <div className="flex items-center justify-between gap-3"><div><p className="text-xs font-extrabold uppercase tracking-[.14em] text-[#6B86A6]">{t("Live preview")}</p><p className="mt-1 text-sm font-bold text-[#171717]">{templates.find(t => t.id === selectedTemplateId)?.name || t("Your CV")}</p></div><Button type="button" onClick={() => switchTemplate(selectedTemplateId)} disabled={!selectedTemplateId || switchingTemplate} className="bg-[#004AAD] text-white hover:bg-[#003B8E] font-bold">{switchingTemplate ? t("Opening…") : t("Continue")}</Button></div>
            <div className="template-choice-paper mt-4"><CvPreview content={cv.content} title={cv.title} template={templates.find(t => t.id === selectedTemplateId)} /></div>
          </aside>
        </div>
      </div>
      <div className={`editor-layout editor-mobile-${mobileEditorView}`}>
        
        {/* LEFT COLUMN: EDITOR CONTROLS & FORM */}
        <section className="editor-controls">
          <div className="flex items-center justify-between border-b border-[#E5E7EB] pb-4">
            <button type="button" onClick={() => { setMobileEditorView("templates"); try { sessionStorage.removeItem("tengenezacv_mobile_editor_view"); } catch {} }} className="mobile-editor-back inline-flex items-center gap-2 text-xs font-bold text-[#004AAD]"><ArrowLeft size={15} /> Change template</button>
            <Link
              href="/workspace"
              className="inline-flex items-center gap-2 text-xs font-bold text-[#525252] hover:text-[#004AAD]"
            >
              <ArrowLeft size={15} /> Workspace
            </Link>
            <Button
              onClick={() => save()}
              disabled={saving}
              className="bg-[#004AAD] text-white hover:bg-[#003B8E] font-bold text-xs"
            >
              {saving ? (
                <Loader2 className="mr-1.5 animate-spin" size={14} />
              ) : (
                <Save className="mr-1.5" size={14} />
              )}
              {t("Save Document")}
            </Button>
          </div>

          <div className="mt-5">
            <input
              type="text"
              value={cv.title}
              onChange={e => setCv({ ...cv, title: e.target.value })}
              className="w-full text-2xl font-extrabold text-[#171717] outline-none border-b border-transparent focus:border-[#004AAD] py-1"
              placeholder={t("CV Document Title")}
            />
          </div>

          <div className="mt-6 space-y-8">
            <Section label={t("Template Design")} className="mobile-template-section">
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                {templates.map(t => (
                  <button
                    key={t.id}
                    onClick={() => switchTemplate(t.id)}
                    disabled={switchingTemplate}
                    className={`rounded-lg border p-2.5 text-left transition ${
                      cv.template_id === t.id
                        ? "border-[#004AAD] bg-[#EAF2FF]"
                        : "border-[#E5E7EB] bg-white hover:border-[#C9D1CC]"
                    }`}
                  >
                    <span className="block text-xs font-bold text-[#171717] truncate">
                      {t.name}
                    </span>
                    <span className="block text-[10px] capitalize text-[#737373]">
                      {t.family.replace("_", " ")}
                    </span>
                  </button>
                ))}
              </div>
            </Section>

            <Section label={t("Personal & Contact")}>
              <div className="space-y-4">
                {/* Profile Photo Upload Control */}
                <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 rounded-lg border border-[#E5E7EB] bg-[#F7F7F7] p-3.5">
                  <div className="relative h-14 w-14 overflow-hidden rounded-full border-2 border-[#004AAD]/20 bg-white flex items-center justify-center text-[#737373] shrink-0 shadow-xs">
                    {personal.photo_url ? (
                      <img
                        src={assetUrl(personal.photo_url)}
                        alt="Profile photo"
                        className="h-full w-full object-cover"
                      />
                    ) : (
                      <User size={24} />
                    )}
                  </div>
                  <div className="flex-1 space-y-1 w-full">
                    <div className="flex items-center justify-between">
                      <b className="text-xs font-bold text-[#171717]">{t("Profile Photo")}</b>
                      {personal.photo_url && (
                        <button
                          type="button"
                          onClick={() => personalChange("photo_url", "")}
                          className="text-xs font-bold text-[#C72F4D] hover:underline"
                        >
                          {t("Remove photo")}
                        </button>
                      )}
                    </div>
                    <p className="text-[11px] text-[#737373]">
                      JPEG, PNG, or WebP up to 5 MB. Uploaded photos are applied to this CV after you save.
                    </p>
                    <div className="flex flex-wrap items-center gap-2 pt-1">
                      <label className="inline-flex cursor-pointer items-center gap-1.5 rounded-lg bg-[#004AAD] px-3 py-1.5 text-xs font-bold text-white transition hover:bg-[#003B8E] shadow-xs">
                        {uploadingPhoto ? (
                          <Loader2 size={13} className="animate-spin" />
                        ) : (
                          <Upload size={13} />
                        )}
                        <span>{uploadingPhoto ? t("Uploading...") : t("Upload Photo")}</span>
                        <input
                          type="file"
                          accept="image/jpeg,image/png,image/webp"
                          className="hidden"
                          onChange={uploadPhoto}
                        />
                      </label>
                      <Input
                        value={personal.photo_url || ""}
                        onChange={e => personalChange("photo_url", e.target.value)}
                        placeholder={t("Or paste photo URL...")}
                        className="h-8 text-xs flex-1 min-w-[160px]"
                      />
                    </div>
                  </div>
                </div>

                <div className="grid gap-3 sm:grid-cols-2">
                  <Field label={t("Full Name")}>
                    <Input
                      value={personal.full_name || ""}
                      onChange={e => personalChange("full_name", e.target.value)}
                      placeholder="e.g. Asha Mollel"
                    />
                  </Field>
                  <Field label={t("Professional Headline")}>
                    <Input
                      value={personal.headline || ""}
                      onChange={e => personalChange("headline", e.target.value)}
                      placeholder="e.g. Operations Coordinator"
                    />
                  </Field>
                </div>
                <div className="grid gap-3 sm:grid-cols-3">
                  <Field label={t("Email")}>
                    <Input
                      value={personal.email || ""}
                      onChange={e => personalChange("email", e.target.value)}
                      placeholder="e.g. asha.mollel@example.com"
                    />
                  </Field>
                  <Field label={t("Phone")}>
                    <Input
                      value={personal.phone || ""}
                      onChange={e => personalChange("phone", e.target.value)}
                      placeholder="e.g. +255 765 990 011"
                    />
                  </Field>
                  <Field label={t("Location")}>
                    <Input
                      value={personal.location || ""}
                      onChange={e => personalChange("location", e.target.value)}
                      placeholder="e.g. Dodoma, Tanzania"
                    />
                  </Field>
                </div>
                <div className="grid gap-3 sm:grid-cols-2">
                  <Field label={t("LinkedIn URL")}>
                    <Input
                      value={personal.linkedin_url || ""}
                      onChange={e => personalChange("linkedin_url", e.target.value)}
                      placeholder="e.g. linkedin.com/in/asha-mollel"
                    />
                  </Field>
                  <Field label={t("Portfolio / Website URL")}>
                    <Input
                      value={personal.portfolio_url || ""}
                      onChange={e => personalChange("portfolio_url", e.target.value)}
                      placeholder="e.g. ashamollel.com"
                    />
                  </Field>
                </div>
              </div>
            </Section>

            <Section label={t("Professional Summary")}>
              <Textarea
                rows={4}
                value={content.summary || ""}
                onChange={e => change({ ...content, summary: e.target.value })}
                placeholder="Briefly state your core value proposition..."
              />
            </Section>

            <Section label={t("Work Experience")}>
              <div className="space-y-4">
                {exp.map((x: any, i: number) => (
                  <div
                    key={i}
                    className="rounded-lg border border-[#E5E7EB] bg-white p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <b className="text-xs font-extrabold text-[#171717]">
                        {t("Role")} {i + 1}
                      </b>
                      <button
                        onClick={() =>
                          change({
                            ...content,
                            experience: exp.filter((_, n) => n !== i),
                          })
                        }
                        className="text-[#C72F4D] hover:text-red-700"
                        title="Remove role"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={t("Title / Role")}>
                        <Input
                          value={roleOf(x)}
                          onChange={e => experienceChange(i, "title", e.target.value)}
                          placeholder="e.g. Finance & Compliance Officer"
                        />
                      </Field>
                      <Field label={t("Employer")}>
                        <Input
                          value={employerOf(x)}
                          onChange={e => experienceChange(i, "employer", e.target.value)}
                          placeholder="e.g. CRDB Bank"
                        />
                      </Field>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-3">
                      <Field label={t("Location")}>
                        <Input
                          value={x.location || ""}
                          onChange={e => experienceChange(i, "location", e.target.value)}
                          placeholder="e.g. Dar es Salaam"
                        />
                      </Field>
                      <Field label={t("Start Date")}>
                        <Input
                          value={x.start_date || ""}
                          onChange={e => experienceChange(i, "start_date", e.target.value)}
                          placeholder="e.g. 2021"
                        />
                      </Field>
                      <Field label={t("End Date")}>
                        <Input
                          value={x.end_date || ""}
                          onChange={e => experienceChange(i, "end_date", e.target.value)}
                          placeholder="e.g. Present"
                        />
                      </Field>
                    </div>
                    <Field label={t("Key Achievements & Bullet Points (one per line)")}>
                      <Textarea
                        rows={4}
                        value={getBulletsText(x)}
                        onChange={e => handleBulletsChange(i, e.target.value)}
                        placeholder="• Led AML transaction-monitoring reviews across 40,000 monthly transactions, reducing false-positive escalations by 28%&#10;• Trained 18 branch colleagues on updated compliance controls&#10;• Automated monthly risk-report consolidation in Power BI"
                      />
                    </Field>
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    change({
                      ...content,
                      experience: [
                        ...exp,
                        { title: "", employer: "", location: "", start_date: "", end_date: "Present", is_current: true, bullets: [] },
                      ],
                    })
                  }
                  className="w-full border-dashed border-[#004AAD]/40 text-[#004AAD]"
                >
                  <Plus size={15} className="mr-1.5" / > {t("Add Experience")}
                </Button>
              </div>
            </Section>

            <Section label={t("Education")}>
              <div className="space-y-4">
                {edu.map((x: any, i: number) => (
                  <div
                    key={i}
                    className="rounded-lg border border-[#E5E7EB] bg-white p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <b className="text-xs font-extrabold text-[#171717]">
                        {t("Education")} {i + 1}
                      </b>
                      <button
                        onClick={() =>
                          change({
                            ...content,
                            education: edu.filter((_, n) => n !== i),
                          })
                        }
                        className="text-[#C72F4D] hover:text-red-700"
                        title="Remove education"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={t("Institution")}>
                        <Input
                          value={x.institution || ""}
                          onChange={e => educationChange(i, "institution", e.target.value)}
                          placeholder="e.g. University of Dar es Salaam"
                        />
                      </Field>
                      <Field label={t("Qualification / Degree")}>
                        <Input
                          value={credentialOf(x)}
                          onChange={e => educationChange(i, "qualification", e.target.value)}
                          placeholder="e.g. Bachelor of Commerce in Finance"
                        />
                      </Field>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={t("Start Date")}>
                        <Input
                          value={x.start_date || ""}
                          onChange={e => educationChange(i, "start_date", e.target.value)}
                          placeholder="e.g. 2017"
                        />
                      </Field>
                      <Field label={t("End Date / Graduated")}>
                        <Input
                          value={x.end_date || ""}
                          onChange={e => educationChange(i, "end_date", e.target.value)}
                          placeholder="e.g. 2020"
                        />
                      </Field>
                    </div>
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    change({
                      ...content,
                      education: [...edu, { institution: "", qualification: "", start_date: "", end_date: "" }],
                    })
                  }
                  className="w-full border-dashed border-[#004AAD]/40 text-[#004AAD]"
                >
                  <Plus size={15} className="mr-1.5" / > {t("Add Education")}
                </Button>
              </div>
            </Section>

            <Section label={t("Skills & Languages")}>
              <div className="space-y-3">
                <Field label={t("Hard Skills (comma separated)")}>
                  <Input
                    value={
                      hardSkillsDraft !== null
                        ? hardSkillsDraft
                        : (skills.hard_skills || []).join(", ")
                    }
                    onChange={e => setHardSkillsDraft(e.target.value)}
                    onBlur={() => {
                      if (hardSkillsDraft !== null) {
                        change({
                          ...content,
                          skills: {
                            ...skills,
                            hard_skills: splitSkills(hardSkillsDraft),
                          },
                        });
                        setHardSkillsDraft(null);
                      }
                    }}
                    placeholder="Financial modelling, Excel, Power BI, Compliance, SQL"
                  />
                </Field>
                <Field label={t("Strengths / Soft Skills (comma separated)")}>
                  <Input
                    value={
                      softSkillsDraft !== null
                        ? softSkillsDraft
                        : (skills.soft_skills || []).join(", ")
                    }
                    onChange={e => setSoftSkillsDraft(e.target.value)}
                    onBlur={() => {
                      if (softSkillsDraft !== null) {
                        change({
                          ...content,
                          skills: {
                            ...skills,
                            soft_skills: splitSkills(softSkillsDraft),
                          },
                        });
                        setSoftSkillsDraft(null);
                      }
                    }}
                    placeholder="Leadership, Analytical Thinking, Audit Readiness"
                  />
                </Field>
                <Field label={t("Languages (comma separated)")}>
                  <Input
                    value={
                      languagesDraft !== null
                        ? languagesDraft
                        : (skills.languages || [])
                            .map((l: any) => (typeof l === "string" ? l : `${l.language} (${l.proficiency})`))
                            .join(", ")
                    }
                    onChange={e => setLanguagesDraft(e.target.value)}
                    onBlur={() => {
                      if (languagesDraft !== null) {
                        change({
                          ...content,
                          skills: {
                            ...skills,
                            languages: splitSkills(languagesDraft),
                          },
                        });
                        setLanguagesDraft(null);
                      }
                    }}
                    placeholder="English (Fluent), Swahili (Native)"
                  />
                </Field>
              </div>
            </Section>

            <Section label={t("Certifications")}>
              <div className="space-y-4">
                {certs.map((c: any, i: number) => (
                  <div
                    key={i}
                    className="rounded-lg border border-[#E5E7EB] bg-white p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <b className="text-xs font-extrabold text-[#171717]">
                        {t("Certification")} {i + 1}
                      </b>
                      <button
                        onClick={() =>
                          change({
                            ...content,
                            certifications: certs.filter((_, n) => n !== i),
                          })
                        }
                        className="text-[#C72F4D] hover:text-red-700"
                        title="Remove certification"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-3">
                      <Field label={t("Certification Name")}>
                        <Input
                          value={c.name || ""}
                          onChange={e => certificationChange(i, "name", e.target.value)}
                          placeholder="e.g. CAMS Certification"
                        />
                      </Field>
                      <Field label={t("Issuer")}>
                        <Input
                          value={c.issuer || ""}
                          onChange={e => certificationChange(i, "issuer", e.target.value)}
                          placeholder="e.g. ACAMS"
                        />
                      </Field>
                      <Field label={t("Year")}>
                        <Input
                          value={c.year || ""}
                          onChange={e => certificationChange(i, "year", e.target.value)}
                          placeholder="e.g. 2023"
                        />
                      </Field>
                    </div>
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    change({
                      ...content,
                      certifications: [...certs, { name: "", issuer: "", year: "" }],
                    })
                  }
                  className="w-full border-dashed border-[#004AAD]/40 text-[#004AAD]"
                >
                  <Plus size={15} className="mr-1.5" / > {t("Add Certification")}
                </Button>
              </div>
            </Section>

            <Section label={t("Referees")}>
              <div className="space-y-4">
                {referees.map((r: any, i: number) => (
                  <div
                    key={i}
                    className="rounded-lg border border-[#E5E7EB] bg-white p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <b className="text-xs font-extrabold text-[#171717]">
                        {t("Referee")} {i + 1}
                      </b>
                      <button
                        onClick={() =>
                          change({
                            ...content,
                            referees: referees.filter((_, n) => n !== i),
                          })
                        }
                        className="text-[#C72F4D] hover:text-red-700"
                        title="Remove referee"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={t("Full Name")}>
                        <Input
                          value={r.name || ""}
                          onChange={e => refereeChange(i, "name", e.target.value)}
                          placeholder="e.g. Jane Doe"
                        />
                      </Field>
                      <Field label={t("Title / Role")}>
                        <Input
                          value={r.title || ""}
                          onChange={e => refereeChange(i, "title", e.target.value)}
                          placeholder="e.g. Head of Compliance"
                        />
                      </Field>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-3">
                      <Field label={t("Organization")}>
                        <Input
                          value={r.organization || ""}
                          onChange={e => refereeChange(i, "organization", e.target.value)}
                          placeholder="e.g. CRDB Bank"
                        />
                      </Field>
                      <Field label={t("Phone")}>
                        <Input
                          value={r.phone || ""}
                          onChange={e => refereeChange(i, "phone", e.target.value)}
                          placeholder="e.g. +255 700 000 000"
                        />
                      </Field>
                      <Field label={t("Email")}>
                        <Input
                          value={r.email || ""}
                          onChange={e => refereeChange(i, "email", e.target.value)}
                          placeholder="e.g. referee@example.com"
                        />
                      </Field>
                    </div>
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    change({
                      ...content,
                      referees: [...referees, { name: "", title: "", organization: "", phone: "", email: "" }],
                    })
                  }
                  className="w-full border-dashed border-[#004AAD]/40 text-[#004AAD]"
                >
                  <Plus size={15} className="mr-1.5" / > {t("Add Referee")}
                </Button>
              </div>
            </Section>
          </div>
          <Button
            type="button"
            onClick={() => setMobileEditorView("export")}
            className="mobile-preview-export mobile-preview-export-end mt-8 bg-[#004AAD] text-white hover:bg-[#003B8E] font-bold"
          >
            <Eye className="mr-2" size={16} /> {t("Preview & export CV")}
          </Button>
        </section>

        {/* RIGHT COLUMN: LIVE FOLIO & TOP ACTION TOOLBAR */}
        <section className="editor-preview bg-[#F7F7F7]">
          <button type="button" onClick={() => setMobileEditorView("editor")} className="mobile-export-back mb-4 inline-flex items-center gap-2 text-sm font-bold text-[#004AAD]"><ArrowLeft size={16} /> {t("Back to editing")}</button>
          <div className="sticky top-[76px] space-y-4">
            
            {/* TOP PREVIEW ACTION TOOLBAR */}
            <div className="relative flex flex-wrap items-center justify-between gap-2.5 rounded-lg border border-[#E5E7EB] bg-white p-3 shadow-xs">
              
              {/* LEFT ACTIONS: COLOR PICKER, ATS REVIEW, HISTORY */}
              <div className="flex items-center gap-2">
                
                {/* 1. Color Picker Swatch Button */}
                <div className="color-picker-control relative" ref={colorPickerRef}>
                  <span className="select-color-label text-[11px] font-extrabold uppercase tracking-wider text-[#6B86A6]">{t("Select color:")}</span>
                  <button
                    type="button"
                    onClick={() => {
                      setShowColorPicker(!showColorPicker);
                      setShowAtsPopover(false);
                      setShowHistoryPopover(false);
                    }}
                    className="group flex items-center gap-1.5 rounded-lg border border-[#E5E7EB] bg-[#F7F7F7] p-1 pr-2 shadow-xs transition hover:border-[#004AAD] hover:bg-white"
                    title={t("Preview an accent color (not saved to the CV)")}
                  >
                    <span
                      className="inline-block h-5 w-5 rounded-md shadow-xs ring-2 ring-white transition transform group-hover:scale-105"
                      style={{ background: activeColor }}
                    />
                    <span className="color-hex text-[11px] font-extrabold font-mono text-[#171717]">{activeColor.toUpperCase()}</span>
                    <Palette size={12} className="text-[#737373] group-hover:text-[#004AAD]" />
                  </button>

                  {/* Color Picker Dropdown Popover */}
                  {showColorPicker && (
                    <div className="absolute left-0 top-10 z-50 w-64 rounded-lg border border-[#E5E7EB] bg-white p-3 shadow-2xl space-y-2.5 animate-in fade-in zoom-in-95">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-black text-[#171717] uppercase tracking-wider">
                          {t("Preview accent color")}
                        </span>
                        <button
                          onClick={() => setShowColorPicker(false)}
                          className="text-[11px] font-bold text-[#004AAD] hover:underline"
                        >
                          {t("Done")}
                        </button>
                      </div>

                      {/* Presets Grid */}
                      <div className="grid grid-cols-6 gap-1">
                        {PRESET_COLORS.map(c => (
                          <button
                            key={c}
                            onClick={() => setCustomAccentColor(c)}
                            className={`h-5 w-full rounded border flex items-center justify-center transition transform hover:scale-105 ${
                              activeColor.toLowerCase() === c.toLowerCase()
                                ? "ring-2 ring-[#004AAD] ring-offset-1 border-transparent shadow-xs"
                                : "border-slate-200"
                            }`}
                            style={{ background: c }}
                            title={c}
                          >
                            {activeColor.toLowerCase() === c.toLowerCase() && (
                              <Check size={10} className="text-white drop-shadow-xs" />
                            )}
                          </button>
                        ))}
                      </div>

                      {/* Custom Color Input */}
                      <div className="flex items-center gap-1.5 pt-2 border-t border-[#EFF5FD]">
                        <input
                          type="color"
                          value={activeColor.startsWith("#") ? activeColor : "#004AAD"}
                          onChange={e => setCustomAccentColor(e.target.value)}
                          className="h-6 w-7 cursor-pointer rounded border-0 bg-transparent p-0"
                          title={t("Pick custom color")}
                        />
                        <input
                          type="text"
                          value={activeColor}
                          onChange={e => setCustomAccentColor(e.target.value)}
                          placeholder="#004AAD"
                          className="color-hex-input w-full rounded-md border border-[#E5E7EB] bg-[#F7F7F7] px-2 py-0.5 text-xs font-mono font-bold text-[#171717] uppercase outline-none focus:border-[#004AAD] focus:bg-white"
                        />
                      </div>

                      {/* Rangi Tanzania color tool */}
                      <a
                        href="https://www.rangitanzania.online/tengeneza"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center justify-center gap-1.5 w-full rounded-lg border border-[#E5E7EB] bg-[#F3F7FD] py-1.5 px-2 text-[11px] font-bold text-[#004AAD] hover:bg-[#EAF2FF] hover:border-[#004AAD] transition shadow-2xs group"
                      >
                        <span>{t("Explore colors at Rangi Tanzania")}</span>
                        <ExternalLink size={12} className="transition transform group-hover:translate-x-0.5" />
                      </a>
                    </div>
                  )}
                </div>

                <button type="button" onClick={() => { setShowMoreActions(!showMoreActions); setShowColorPicker(false); }} className="preview-more-button flex items-center justify-center rounded-lg border border-[#E5E7EB] bg-white px-2 py-1.5 text-[#171717] shadow-xs hover:border-[#004AAD]" aria-label={t("More CV tools")}>
                  <MoreHorizontal size={17} />
                </button>
                <div className={`preview-secondary-actions ${showMoreActions ? "preview-secondary-actions-open" : ""}`}>
                {/* 2. ATS Review Button */}
                <div className="relative" ref={atsRef}>
                  <button
                    type="button"
                    onClick={() => {
                      setShowAtsPopover(!showAtsPopover);
                      setShowColorPicker(false);
                      setShowHistoryPopover(false);
                      if (!report) runScore();
                    }}
                    className={`flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs font-bold transition shadow-xs ${
                      report
                        ? "border-[#004AAD]/30 bg-[#EAF2FF] text-[#004AAD]"
                        : "border-[#E5E7EB] bg-[#F7F7F7] text-[#171717] hover:border-[#004AAD]"
                    }`}
                  >
                    <Sparkles size={13} className="text-[#004AAD]" />
                    <span>{t("ATS rules check")}</span>
                    {report && (
                      <span className="rounded-full bg-[#004AAD] px-1.5 py-0.2 text-[10px] font-black text-white">
                        {report.score}/100
                      </span>
                    )}
                  </button>

                  {/* ATS Review Popover */}
                  {showAtsPopover && (
                    <div className="ats-popover absolute left-0 top-10 z-50 w-80 rounded-lg border border-[#E5E7EB] bg-white p-4 shadow-2xl space-y-3 animate-in fade-in zoom-in-95">
                      <div className="flex items-center justify-between border-b border-[#EFF5FD] pb-2">
                        <div>
                          <span className="text-[10px] font-black text-[#171717] uppercase tracking-wider">
                            {t("ATS formatting checks")}
                          </span>
                          <p className="text-xs font-black text-[#004AAD]">
                            {report ? `${report.score} / 100 · rule-based score` : t("Checking formatting…")}
                          </p>
                        </div>
                        <Button
                          size="sm"
                          onClick={runScore}
                          className="h-7 text-xs bg-[#004AAD] text-white hover:bg-[#003B8E]"
                        >
                          {t("Run check")}
                        </Button>
                      </div>

                      {report ? (
                        <div className="space-y-2 max-h-60 overflow-y-auto pr-1">
                          {report.breakdown.map((x, i) => (
                            <div
                              key={i}
                              className="rounded-lg border border-[#E5E7EB] bg-[#F7F7F7] p-2.5 text-xs"
                            >
                              <p className="flex items-start gap-2 text-[#525252]">
                                <AlertTriangle
                                  className={
                                    x.severity === "warning"
                                      ? "shrink-0 text-amber-500 mt-0.5"
                                      : "shrink-0 text-[#004AAD] mt-0.5"
                                  }
                                  size={14}
                                />
                                <span>{x.message}</span>
                              </p>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="py-3 flex justify-center">
                          <Loader2 className="animate-spin text-[#004AAD]" size={20} />
                        </div>
                      )}
                      <p className="text-[11px] leading-4 text-[#737373]">Checks document structure and common formatting issues; it does not predict an employer’s screening result.</p>
                    </div>
                  )}
                </div>

                {/* 3. Document History Button */}
                <div className="relative" ref={historyRef}>
                  <button
                    type="button"
                    onClick={() => {
                      setShowHistoryPopover(!showHistoryPopover);
                      setShowColorPicker(false);
                      setShowAtsPopover(false);
                      if (!history) loadHistory();
                    }}
                    className="flex items-center gap-1.5 rounded-lg border border-[#E5E7EB] bg-[#F7F7F7] px-2.5 py-1 text-xs font-bold text-[#171717] shadow-xs transition hover:border-[#004AAD] hover:bg-white"
                  >
                    <FileClock size={13} className="text-[#737373]" />
                    <span>{t("History")}</span>
                  </button>

                  {/* Version History Popover */}
                  {showHistoryPopover && (
                    <div className="history-popover absolute left-0 top-10 z-50 w-72 rounded-lg border border-[#E5E7EB] bg-white p-3.5 shadow-2xl space-y-2.5 animate-in fade-in zoom-in-95">
                      <div className="flex items-center justify-between border-b border-[#EFF5FD] pb-2">
                        <span className="text-[10px] font-black text-[#171717] uppercase tracking-wider">
                          {t("Document History")}
                        </span>
                        <button
                          onClick={() => setShowHistoryPopover(false)}
                          className="text-[11px] font-bold text-[#004AAD] hover:underline"
                        >
                          {t("Close")}
                        </button>
                      </div>

                      {history ? (
                        <div className="space-y-1.5 max-h-60 overflow-y-auto pr-1">
                          {history.length ? (
                            history.map((v, i) => (
                              <button
                                key={v.id || i}
                                onClick={() => restore(v)}
                                className="w-full rounded-lg border border-[#E5E7EB] bg-[#F7F7F7] p-2 text-left transition hover:border-[#004AAD] hover:bg-white"
                              >
                                <b className="block text-xs text-[#171717]">
                                  {v.change_summary || t("Saved version")}
                                </b>
                                <small className="block text-[10px] text-[#004AAD] font-bold mt-0.5">
                                  {t("Click to restore state")}
                                </small>
                              </button>
                            ))
                          ) : (
                            <p className="text-xs text-[#737373]">{t("No saved versions yet.")}</p>
                          )}
                        </div>
                      ) : (
                        <div className="py-3 flex justify-center">
                          <Loader2 className="animate-spin text-[#004AAD]" size={20} />
                        </div>
                      )}
                      <p className="text-[11px] leading-4 text-[#737373]">Checks document structure and common formatting issues; it does not predict an employer’s screening result.</p>
                    </div>
                  )}
                </div>
                </div>
              </div>

              {/* RIGHT ACTIONS: EXPORT BUTTONS */}
              <div className="export-actions flex items-center gap-1.5">
                <span className="download-label text-[11px] font-extrabold uppercase tracking-wider text-[#6B86A6]">{t("Download:")}</span>
                <Button
                  onClick={() => exportCV("csv")}
                  disabled={Boolean(exportingFormat)}
                  size="sm"
                  variant="outline"
                  className="border-[#E5E7EB] text-[#171717] hover:bg-[#F5F5F5] font-bold text-xs h-8"
                >
                  {exportingFormat === "csv" ? <Loader2 className="mr-1.5 animate-spin" size={13} /> : <Download className="mr-1.5" size={13} />}
                  {exportingFormat === "csv" ? t("Preparing…") : "CSV"}
                </Button>
                <Button
                  onClick={() => exportCV("pdf")}
                  disabled={Boolean(exportingFormat)}
                  size="sm"
                  className="bg-[#004AAD] text-white hover:bg-[#003B8E] font-bold text-xs shadow-xs px-3 py-1 h-8"
                >
                  {exportingFormat === "pdf" ? <Loader2 className="mr-1.5 animate-spin" size={13} /> : <Download className="mr-1.5" size={13} />}
                  {exportingFormat === "pdf" ? t("Preparing…") : "PDF"}
                </Button>
                <Button
                  onClick={() => exportCV("docx")}
                  disabled={Boolean(exportingFormat)}
                  size="sm"
                  variant="outline"
                  className="border-[#E5E7EB] text-[#171717] hover:bg-[#F5F5F5] font-bold text-xs h-8"
                >
                  {exportingFormat === "docx" ? <Loader2 className="mr-1.5 animate-spin" size={13} /> : <Download className="mr-1.5" size={13} />}
                  {exportingFormat === "docx" ? t("Preparing…") : "DOCX"}
                </Button>
                <button type="button" onClick={() => setShowEmailPanel(!showEmailPanel)} className={`email-export-trigger inline-flex h-8 items-center justify-center rounded-lg border px-2.5 text-xs font-bold shadow-xs transition ${showEmailPanel ? "border-[#004AAD] bg-[#EAF2FF] text-[#004AAD]" : "border-[#E5E7EB] bg-white text-[#171717] hover:border-[#004AAD]"}`} aria-label={t("Email CV")}><Mail size={14} /></button>
              </div>
            </div>
            <div className={`export-email-card rounded-lg border border-[#DCE9FF] bg-[#F7FAFF] p-3 ${showEmailPanel ? "export-email-card-open" : ""}`}>
              <label htmlFor="delivery-email" className="block text-xs font-extrabold text-[#102A56]">Email me a copy</label>
              <div className="mt-2 flex gap-2">
                <input id="delivery-email" type="email" value={deliveryEmail} onChange={e => setDeliveryEmail(e.target.value)} placeholder="you@example.com" className="min-w-0 flex-1 rounded-lg border border-[#C9D1CC] bg-white px-3 py-2 text-sm outline-none focus:border-[#004AAD]" />
                <Button type="button" disabled={!deliveryEmail.includes("@") || !lastExportFilename || emailSending} onClick={emailLatestExport} className="bg-[#004AAD] text-white hover:bg-[#003B8E]">{emailSending ? "Sending…" : "Send"}</Button>
              </div>
              <p className="mt-1.5 text-[11px] text-[#6B86A6]">Export the CV first, then enter your email to receive it.</p>
            </div>

            {/* CV LIVE DOCUMENT PREVIEW CANVAS */}
            <div className="editor-folio-canvas rounded-lg border border-[#E5E7EB] bg-white p-3 shadow-xs overflow-hidden flex justify-center">
              <div className="editor-folio-scale">
                <CvPreview
                content={cv.content}
                title={cv.title}
                template={
                  activeTemplate
                    ? { ...activeTemplate, accent_color_hex: activeColor }
                    : undefined
                }
                />
              </div>
            </div>

          </div>
        </section>

      </div>
    </AppShell>
  );
}
