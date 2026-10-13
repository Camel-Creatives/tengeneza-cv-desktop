'use client';

/** Ultra-pure 1-page landing page with frameless, scroll-free live CV slideshow. */
import { CvPreview } from "@/components/CvPreview";
import { PublicHeader } from "@/components/AppShell";
import { BrandMark } from "@/components/BrandMark";
import { Button } from "@/components/ui/button";
import { api, ExampleSummary, Template } from "@/lib/api";
import { ArrowRight, Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import Link from "next/link";

export default function Home() {
  const [examples, setExamples] = useState<ExampleSummary[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [docsBySlug, setDocsBySlug] = useState<
    Record<string, { title: string; template_id: string; content: any }>
  >({});
  const [activeIndex, setActiveIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [fade, setFade] = useState(true);

  useEffect(() => {
    Promise.all([api.templatesPublic(), api.examples()])
      .then(([tList, eList]) => {
        setTemplates(tList);
        setExamples(eList);

        // Preload sample documents
        Promise.all(
          eList.map(x =>
            api
              .example(x.slug)
              .then(d => [x.slug, d] as const)
              .catch(() => null)
          )
        ).then(entries => {
          const map: typeof docsBySlug = {};
          for (const entry of entries) if (entry) map[entry[0]] = entry[1];
          setDocsBySlug(map);
        });
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  // Continuous seamless auto-slide loop
  useEffect(() => {
    if (!examples.length) return;
    const timer = setInterval(() => {
      setFade(false);
      setTimeout(() => {
        setActiveIndex(prev => (prev + 1) % examples.length);
        setFade(true);
      }, 250);
    }, 3800);
    return () => clearInterval(timer);
  }, [examples]);

  const currentExample = examples.length ? examples[activeIndex % examples.length] : null;
  const currentDoc = currentExample ? docsBySlug[currentExample.slug] : null;
  const currentTemplate = currentExample
    ? templates.find(t => t.id === currentExample.template_id)
    : null;

  return (
    <div className="min-h-screen bg-[#F7FAFF] flex flex-col font-sans antialiased overflow-x-hidden">
      <PublicHeader />

      {/* ONE PAGE HORIZONTAL HERO & PURE FRAMELESS SLIDESHOW */}
      <main className="flex-1 flex items-center justify-center mx-auto max-w-7xl w-full px-5 py-4 lg:px-8">
        <div className="grid gap-8 lg:grid-cols-[.9fr_1.1fr] items-center w-full">
          
          {/* LEFT HERO COLUMN */}
          <div className="space-y-6">
            <h1 className="text-4xl font-extrabold tracking-tight text-[#102A56] sm:text-5xl lg:text-6xl leading-[1.08]">
              <span className="block">Actual examples.</span>
              <span className="block">See them,</span>
              <span className="block text-[#004AAD]">then make one yours.</span>
            </h1>

            <p className="text-base leading-7 text-[#48617D] max-w-xl">
              These are complete, professionally written sample CVs from the
              Tengeneza library—rendered live in their export designs. Pick one and
              it becomes your editable starting point.
            </p>

            <div className="flex flex-wrap items-center gap-3 pt-2">
              <Link href="/workspace">
                <Button
                  size="lg"
                  className="bg-[#004AAD] text-white shadow-lg shadow-[#004AAD]/20 hover:bg-[#003B8E] px-7 py-5 font-bold"
                >
                  Start Your CV Free <ArrowRight className="ml-2" size={18} />
                </Button>
              </Link>
              <Link href="/templates">
                <Button
                  size="lg"
                  variant="outline"
                  className="border-[#C8DCF8] bg-white text-[#102A56] hover:bg-[#EAF2FF] px-6 py-5 font-bold"
                >
                  All Templates ({templates.length})
                </Button>
              </Link>
            </div>
          </div>

          {/* RIGHT PURE FRAMELESS PASSIVE LIVE SLIDESHOW */}
          <div className="flex items-center justify-center w-full min-h-[580px]">
            {loading || !currentExample ? (
              <div className="flex h-[560px] items-center justify-center">
                <Loader2 className="animate-spin text-[#004AAD]" size={32} />
              </div>
            ) : (
              <div className="h-[580px] w-full flex items-start justify-center pt-2 overflow-hidden pointer-events-none select-none">
                <div
                  className={`w-full flex justify-center transition-all duration-300 transform ${
                    fade ? "opacity-100 scale-100" : "opacity-0 scale-95"
                  }`}
                >
                  <div className="origin-top scale-[0.48] sm:scale-[0.52] lg:scale-[0.55] shadow-2xl rounded-md">
                    {currentDoc ? (
                      <CvPreview
                        content={currentDoc.content}
                        title={currentDoc.title}
                        template={currentTemplate || undefined}
                      />
                    ) : (
                      <div className="flex h-96 w-[600px] items-center justify-center">
                        <Loader2 className="animate-spin text-[#004AAD]" size={24} />
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>

        </div>
      </main>

      <footer className="border-t border-[#DCE9FF] bg-white py-4">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-3 px-5 sm:flex-row lg:px-8">
          <BrandMark compact />
          <p className="text-xs text-[#6B86A6]">
            Tengeneza CV — Professional Watermark-Free Open Workspace
          </p>
        </div>
      </footer>
    </div>
  );
}
