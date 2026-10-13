'use client';

/** Dedicated Example Documents page powered by live Tengeneza API samples with perfect 100% single-screen document fit & interactive Color Picker. */
import { AppShell } from "@/components/AppShell";
import { CvPreview } from "@/components/CvPreview";
import { Button } from "@/components/ui/button";
import { api, ExampleSummary, Template } from "@/lib/api";
import { getToken, useAuthGuard } from "@/lib/auth";
import { useLanguage } from "@/contexts/LanguageContext";
import { ArrowLeft, Copy, ExternalLink, Loader2, Palette, Sparkles, Check } from "lucide-react";
import { useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";

const PRESET_COLORS = [
  "#004AAD", // Tengeneza cobalt
  "#003B8E", // Deep cobalt
  "#2563EB", // Bright blue
  "#5B8DEF", // Clear blue
  "#759FEA", // Soft blue
  "#9DBCF2", // Pale blue
];

export function ExamplesPage() {
  const router = useRouter();
  const { t, locale } = useLanguage();
  const params = useParams();
  const initialSlug = (params?.slug as string) || null;

  const [templates, setTemplates] = useState<Template[]>([]);
  const [examples, setExamples] = useState<ExampleSummary[]>([]);
  const [docsBySlug, setDocsBySlug] = useState<
    Record<string, { title: string; template_id: string; content: any }>
  >({});
  const [loading, setLoading] = useState(true);
  const [activeSlug, setActiveSlug] = useState<string | null>(initialSlug);
  const [mobileView, setMobileView] = useState<"selection" | "preview">(initialSlug ? "preview" : "selection");
  const [loadingExample, setLoadingExample] = useState(false);
  const [using, setUsing] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [mobilePreviewZoom, setMobilePreviewZoom] = useState<number | null>(null);
  const previewFrameRef = useRef<HTMLDivElement>(null);
  const token = getToken();
  useAuthGuard();

  // Interactive Accent Color Picker state
  const [customColor, setCustomColor] = useState<string | null>(null);
  const [showColorPicker, setShowColorPicker] = useState(false);
  const pickerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setMobileView(initialSlug ? "preview" : "selection");
  }, [initialSlug]);

  useEffect(() => {
    Promise.all([api.templatesPublic(), api.examples()])
      .then(([t, e]) => {
        setTemplates(t);
        setExamples(e);
        const sel = initialSlug && e.some(x => x.slug === initialSlug) ? initialSlug : (e.length ? e[0].slug : null);
        setActiveSlug(sel);
        if (sel) {
          api.example(sel).then(d => {
            setDocsBySlug(prev => ({ ...prev, [sel]: d }));
          }).catch(() => {});
        }
      })
      .catch(err =>
        toast.error(
          err instanceof Error ? err.message : "Could not load examples."
        )
      )
      .finally(() => setLoading(false));
  }, [initialSlug]);

  useEffect(() => {
    if (!activeSlug || docsBySlug[activeSlug]) return;
    setLoadingExample(true);
    api
      .example(activeSlug)
      .then(doc => setDocsBySlug(x => ({ ...x, [activeSlug]: doc })))
      .catch(() => toast.error("Example document unavailable."))
      .finally(() => setLoadingExample(false));
  }, [activeSlug, docsBySlug]);

  // Reset custom color on active slug change
  useEffect(() => {
    setCustomColor(null);
  }, [activeSlug]);

  // Close color picker on click outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        setShowColorPicker(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    const frame = previewFrameRef.current;
    if (!frame || typeof ResizeObserver === "undefined") return;

    const updatePreviewZoom = () => {
      if (window.innerWidth > 767) {
        setMobilePreviewZoom(null);
        return;
      }

      // The folio and its padding occupy about 705 CSS pixels at desktop size.
      setMobilePreviewZoom(Math.min(0.72, Math.max(0.38, (frame.clientWidth - 12) / 705)));
    };

    const observer = new ResizeObserver(updatePreviewZoom);
    observer.observe(frame);
    updatePreviewZoom();
    return () => observer.disconnect();
  }, [mobileView, loading]);

  async function cloneExample(template_id: string, content: any, title?: string) {
    setUsing(template_id);
    try {
      const cv = await api.createCV(token, {
        title: title || "Cloned CV Example",
        template_id,
        document_style: "resume",
        target_region: "eac",
        language: locale,
        content,
      });
      toast.success("Example cloned to your workspace!");
      router.push(`/editor/${cv.id}`);
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Could not clone this example."
      );
    } finally {
      setUsing(null);
    }
  }

  const exampleDoc = activeSlug ? docsBySlug[activeSlug] : null;
  const activeTemplate = templates.find(
    t => t.id === exampleDoc?.template_id
  );

  const activeColor = customColor || activeTemplate?.accent_color_hex || "#004AAD";

  const filteredExamples = examples.filter(x => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      x.title.toLowerCase().includes(q) ||
      (x.headline && x.headline.toLowerCase().includes(q)) ||
      (x.template_id && x.template_id.toLowerCase().includes(q))
    );
  });

  return (
    <AppShell>
      <div className="mx-auto flex h-[calc(100dvh-65px)] max-w-7xl flex-col overflow-hidden px-3 pb-3 pt-3 sm:px-5 lg:px-8">
        
        {/* COMPACT TOP HEADER */}
        <div className="flex-none border-b border-[#E5E7EB] pb-2.5">
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-1.5">
                <Sparkles className="text-[#004AAD]" size={15} />
                <p className="ledger-label text-[10px]">
                  <span></span> {t("EXAMPLE DOCUMENTS FROM THE API")}
                </p>
              </div>
              <h1 className="mt-0.5 text-lg font-extrabold text-[#171717] sm:text-xl">
                {t("See finished CVs built on these designs")}
              </h1>
              {mobileView === "preview" && (
                <button type="button" onClick={() => { setMobileView("selection"); router.push("/examples"); }} className="mt-2 inline-flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs font-bold text-[#004AAD] hover:bg-[#EAF2FF] md:hidden">
                  <ArrowLeft size={15} /> {t("Back to example selection")}
                </button>
              )}
            </div>
          </div>
        </div>

        {loading ? (
          <div className="flex-1 flex items-center justify-center">
            <Loader2 className="animate-spin text-[#004AAD]" size={32} />
          </div>
        ) : (
          <div className="flex-1 mt-3 grid gap-5 lg:grid-cols-[290px_1fr] overflow-hidden min-h-0">
            
            {/* LEFT SIDEBAR EXAMPLES LIST */}
            <aside className={`${mobileView === "selection" ? "flex" : "hidden"} flex-col space-y-2 overflow-hidden min-h-0 md:flex`}>
              <input
                type="text"
                placeholder={t("Search examples by role...")}
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                className="w-full flex-none rounded-lg border border-[#E5E7EB] bg-white px-3 py-1.5 text-xs text-[#171717] outline-none focus:border-[#004AAD]"
              />
              <div className="hidden flex-1 space-y-1.5 overflow-y-auto pr-1 md:block">
                {filteredExamples.map(x => (
                  <button
                    key={x.slug}
                    onClick={() => {
                      setActiveSlug(x.slug);
                      setMobileView("preview");
                      router.push(`/examples/${x.slug}`);
                    }}
                    className={`w-full rounded-lg border p-2.5 text-left transition ${
                      activeSlug === x.slug
                        ? "border-[#004AAD] bg-[#EAF2FF] shadow-xs"
                        : "border-[#E5E7EB] bg-white hover:border-[#C9D1CC]"
                    }`}
                  >
                    <b className="block text-xs font-bold text-[#171717] truncate">
                      {x.title}
                    </b>
                    <small className="mt-0.5 block text-[10px] capitalize text-[#737373] truncate">
                      {(x.template_id || "").replaceAll("_", " ")}
                      {x.headline ? ` · ${x.headline}` : ""}
                    </small>
                  </button>
                ))}
              </div>
              <div className="min-h-0 flex-1 overflow-auto rounded-lg border border-[#E5E7EB] bg-white md:hidden">
                <table className="w-full min-w-[470px] border-collapse text-left">
                  <thead className="sticky top-0 z-10 bg-[#F3F7FD] text-[10px] font-extrabold uppercase tracking-wide text-[#48617D]">
                    <tr><th className="px-3 py-2.5">{t("Example")}</th><th className="px-3 py-2.5">{t("Role")}</th><th className="px-3 py-2.5">{t("Template")}</th></tr>
                  </thead>
                  <tbody className="divide-y divide-[#E5E7EB]">
                    {filteredExamples.map(x => (
                      <tr key={x.slug} onClick={() => { setActiveSlug(x.slug); setMobileView("preview"); router.push(`/examples/${x.slug}`); }} role="button" tabIndex={0} aria-label={`Open ${x.title}`} onKeyDown={event => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); setActiveSlug(x.slug); setMobileView("preview"); router.push(`/examples/${x.slug}`); } }} className="cursor-pointer transition hover:bg-[#F7FAFF] focus-visible:outline focus-visible:outline-2 focus-visible:outline-[#004AAD] active:bg-[#EAF2FF]">
                        <td className="max-w-[150px] px-3 py-3 text-xs font-bold text-[#171717]">{x.title}</td>
                        <td className="max-w-[190px] px-3 py-3 text-xs text-[#525252]">{x.headline || "—"}</td>
                        <td className="whitespace-nowrap px-3 py-3 text-[11px] font-semibold capitalize text-[#48617D]">{(x.template_id || "").replaceAll("_", " ")}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {!filteredExamples.length && <p className="p-4 text-sm text-[#737373]">{t("No examples match your search.")}</p>}
              </div>
            </aside>

            {/* RIGHT DOCUMENT DISPLAY & COLOR PICKER */}
            <div className={`${mobileView === "preview" ? "flex" : "hidden"} flex-col overflow-hidden min-h-0 md:flex`}>
              {activeTemplate && exampleDoc && (
                <div className="flex-none mb-2.5 flex flex-col items-stretch justify-between gap-2 rounded-lg border border-[#E5E7EB] bg-white px-3 py-2 shadow-xs sm:flex-row sm:items-center sm:px-3.5">
                  
                  {/* DESIGN SYSTEM & INTERACTIVE COLOR PICKER */}
                  <div className="flex min-w-0 items-start gap-2.5 sm:items-center">
                    
                    {/* Interactive Color Picker Swatch Button */}
                    <div className="relative" ref={pickerRef}>
                      <button
                        type="button"
                        onClick={() => setShowColorPicker(!showColorPicker)}
                        className="group flex items-center gap-1.5 rounded-lg border border-[#E5E7EB] bg-[#F7F7F7] p-1 pr-2 shadow-xs transition hover:border-[#004AAD] hover:bg-white"
                        title={t("Click to select or type accent color")}
                      >
                        <span
                          className="inline-block h-5 w-5 rounded-md shadow-xs ring-2 ring-white transition transform group-hover:scale-105"
                          style={{ background: activeColor }}
                        />
                        <span className="text-[11px] font-extrabold text-[#171717] font-mono">
                          {activeColor.toUpperCase()}
                        </span>
                        <Palette size={12} className="text-[#737373] transition group-hover:text-[#004AAD]" />
                      </button>

                      {/* Color Picker Dropdown Popover */}
                      {showColorPicker && (
                        <div className="absolute left-0 top-10 z-50 w-64 rounded-lg border border-[#E5E7EB] bg-white p-3 shadow-2xl space-y-2.5 animate-in fade-in zoom-in-95">
                          <div className="flex items-center justify-between">
                            <span className="text-[10px] font-black text-[#171717] uppercase tracking-wider">
                              {t("Accent Palette")}
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
                                onClick={() => setCustomColor(c)}
                                className={`h-5 w-full rounded transition transform hover:scale-105 border flex items-center justify-center ${
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
                              onChange={e => setCustomColor(e.target.value)}
                              className="h-6 w-7 cursor-pointer rounded border-0 bg-transparent p-0"
                              title={t("Pick custom color")}
                            />
                            <input
                              type="text"
                              value={activeColor}
                              onChange={e => setCustomColor(e.target.value)}
                              placeholder="#004AAD"
                              className="w-full rounded-md border border-[#E5E7EB] bg-[#F7F7F7] px-2 py-0.5 text-xs font-mono font-bold text-[#171717] uppercase outline-none focus:border-[#004AAD] focus:bg-white"
                            />
                          </div>

                          {/* Rangi Tanzania Color Link Redirect */}
                          <a
                            href="https://www.rangitanzania.online/tengeneza"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center justify-center gap-1.5 w-full rounded-lg border border-[#E5E7EB] bg-[#F3F7FD] py-1.5 px-2 text-[11px] font-bold text-[#004AAD] hover:bg-[#EAF2FF] hover:border-[#004AAD] transition shadow-2xs group"
                          >
                            <span>{t("Get More Colors at Rangi Tanzania")}</span>
                            <ExternalLink size={12} className="transition transform group-hover:translate-x-0.5" />
                          </a>
                        </div>
                      )}
                    </div>

                    <div className="min-w-0">
                      <b className="block text-xs font-extrabold text-[#171717]">
                        {t("Rendered as {name}").replace("{name}", activeTemplate.name)}
                      </b>
                      <p className="text-[10px] text-[#737373]">
                        {t("Click color swatch to customize accent theme live.")}
                      </p>
                    </div>
                  </div>

                  <Button
                    onClick={() =>
                      cloneExample(
                        exampleDoc.template_id,
                        exampleDoc.content,
                        exampleDoc.title
                      )
                    }
                    disabled={!!using}
                    size="sm"
                    className="w-full bg-[#004AAD] text-white hover:bg-[#003B8E] font-bold shadow-xs text-xs px-3.5 py-1 h-8 sm:w-auto"
                  >
                    {using ? (
                      <Loader2 className="mr-1.5 animate-spin" size={13} />
                    ) : (
                      <Copy className="mr-1.5" size={13} />
                    )}
                    {t("Clone this example")}
                  </Button>
                </div>
              )}

              {/* DOCUMENT VIEW CONTAINER FIT WITHOUT PAGE SCROLLING */}
              <div ref={previewFrameRef} className="example-document-frame flex-1 rounded-lg border border-[#E5E7EB] bg-white p-1.5 shadow-xs overflow-hidden flex items-center justify-center min-h-0">
                {loadingExample || !exampleDoc ? (
                  <div className="flex h-full items-center justify-center">
                    <Loader2 className="animate-spin text-[#004AAD]" size={28} />
                  </div>
                ) : (
                  <div className="example-document-inner h-full w-full flex justify-center items-center overflow-hidden py-1">
                    <div className="example-preview-paper shadow-md rounded-md transition-transform duration-200" style={mobilePreviewZoom === null ? undefined : { zoom: mobilePreviewZoom }}>
                      <CvPreview
                        content={exampleDoc.content}
                        title={exampleDoc.title}
                        template={{
                          ...activeTemplate,
                          accent_color_hex: activeColor,
                        }}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>

          </div>
        )}
      </div>
    </AppShell>
  );
}
