"use client";

/** Tengeneza CV workspace shell with compact navigation for document work. */
import { BrandMark } from "@/components/BrandMark";
import { Button } from "@/components/ui/button";
import { useLanguage } from "@/contexts/LanguageContext";
import { getUser } from "@/lib/auth";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  ShieldCheck,
  PanelLeftClose,
  PanelLeftOpen,
  LayoutTemplate,
  HardDrive,
  Menu,
  Sparkles,
  UserRound,
  X,
  ChevronRight,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";

const links = [
  { href: "/workspace", icon: LayoutDashboard, label: "workspace" as const },
  { href: "/templates", icon: LayoutTemplate, label: "templates" as const },
  { href: "/examples", icon: Sparkles, label: "examples" as const },
];

function LocaleControl() {
  const { locale, setLocale, t } = useLanguage();
  return (
    <div
      aria-label={t("language")}
      className="flex gap-1 rounded-md border border-[#E5E7EB] bg-[#F7F7F7] p-1"
    >
      <button
        onClick={() => setLocale("en")}
        aria-label={t("english")}
        title={t("english")}
        className={cn(
          "locale-button",
          locale === "en" && "locale-button-active"
        )}
      >
        EN
      </button>
      <button
        onClick={() => setLocale("sw")}
        className={cn(
          "locale-button",
          locale === "sw" && "locale-button-active"
        )}
        aria-label={t("swahili")}
        title={t("swahili")}
      >
        SW
      </button>
    </div>
  );
}

export function AppShell({
  children,
}: {
  children: React.ReactNode;
  onSignOut?: () => void;
}) {
  const location = usePathname();
  const [open, setOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const [isSuperAdmin, setIsSuperAdmin] = useState(false);
  const { t } = useLanguage();

  useEffect(() => {
    setIsSuperAdmin(getUser()?.role === "superadmin");
    try {
      setCollapsed(
        window.localStorage.getItem("tengenezacv_sidebar_collapsed") === "true"
      );
    } catch {
      // The navigation remains usable when browser storage is unavailable.
    }
  }, []);

  const compact = collapsed && !open;
  const pageName = location === "/admin"
    ? t("administration")
    : location.startsWith("/editor")
    ? t("cvEditor")
    : location === "/workspace"
      ? t("workspace")
      : location.startsWith("/templates")
        ? t("templates")
        : location.startsWith("/examples")
          ? t("examples")
          : t("profile");

  const nav = (
    <nav
      aria-label={t("Main navigation")}
      className="flex h-full flex-col justify-between p-3"
    >
      <div>
        <Link
          href="/workspace"
          aria-label="Tengeneza CV workspace"
          onClick={() => setOpen(false)}
          className="block rounded-md focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#004AAD]"
        >
          <BrandMark
            className={cn(
              "sidebar-brand px-3 py-4",
              compact && "justify-center px-0 [&>span]:hidden"
            )}
          />
        </Link>
        <div
          className={cn(
            "mb-2 mt-8 px-3 text-[11px] font-semibold text-[#737373]",
            compact && "sr-only"
          )}
        >
          {t("careerWorkspace")}
        </div>
        <div className="space-y-1">
          {links.map(({ href, icon: Icon, label }) => (
            <Link
              key={label}
              href={href}
              title={compact ? t(label) : undefined}
              aria-label={t(label)}
              onClick={() => setOpen(false)}
              className={cn(
                "nav-link",
                compact && "justify-center px-0",
                (location === href ||
                  (label === "workspace" && location.startsWith("/editor"))) &&
                  "nav-link-active"
              )}
            >
              <Icon size={17} />
              <span className={cn(compact && "sr-only")}>{t(label)}</span>
            </Link>
          ))}
          {isSuperAdmin && (
            <Link
              href="/admin"
              onClick={() => setOpen(false)}
              aria-label={t("administration")}
              className={cn("nav-link", compact && "justify-center px-0", location === "/admin" && "nav-link-active")}
            >
              <ShieldCheck size={17} />
              <span className={cn(compact && "sr-only")}>{t("administration")}</span>
            </Link>
          )}
        </div>
      </div>
      <div className="space-y-2 border-t border-[#E5E7EB] px-2 pt-3">
        <Link
          href="/profile"
          title={compact ? t("profile") : undefined}
          aria-label={t("profile")}
          className={cn(
            "nav-link",
            compact && "justify-center px-0",
            location === "/profile" && "nav-link-active"
          )}
          onClick={() => setOpen(false)}
        >
          <UserRound size={17} />
          <span className={cn("min-w-0 flex-1", compact && "sr-only")}>
            {t("profile")}
          </span>
          <ChevronRight
            size={15}
            className={cn("text-[#A3A3A3]", compact && "hidden")}
          />
        </Link>
        <div
          className={cn(
            "flex items-center justify-between px-2 py-2",
            compact && "hidden"
          )}
        >
          <span className="text-xs font-medium text-[#737373]">{t("language")}</span>
          <LocaleControl />
        </div>
        <div
          title={`${t("Local workspace")} — ${t("Unlimited exports")}`}
          className={cn("flex items-center gap-2 rounded-md bg-[#F4F8FF] px-2 py-2", compact && "justify-center")}
        >
          <HardDrive size={15} className="shrink-0 text-[#004AAD]" />
          <span className={cn("text-[11px] leading-4 text-[#525252]", compact && "sr-only")}>
            {t("Local workspace")}<br />{t("Unlimited exports")}
          </span>
        </div>
      </div>
    </nav>
  );

  function toggleSidebar() {
    const next = !collapsed;
    setCollapsed(next);
    try {
      window.localStorage.setItem(
        "tengenezacv_sidebar_collapsed",
        String(next)
      );
    } catch {
      // A storage failure should not block the current navigation toggle.
    }
  }

  return (
    <div className="app-shell min-h-screen bg-[#F7F7F7] text-[#171717]">
      <a href="#main-content" className="skip-link">
        {t("Skip to content")}
      </a>
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-30 hidden border-r border-[#E5E7EB] bg-white transition-[width] duration-200 lg:block",
          collapsed ? "w-[72px]" : "w-[252px]"
        )}
      >
        {nav}
      </aside>
      <div
        className={cn(
          "transition-[padding] duration-200",
          collapsed ? "lg:pl-[72px]" : "lg:pl-[252px]"
        )}
      >
        <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-[#E5E7EB] bg-white px-4 sm:px-6">
          <div className="flex min-w-0 items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className="text-[#525252] lg:hidden"
              onClick={() => setOpen(true)}
              aria-label={t("Open navigation menu")}
            >
              <Menu size={19} />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="hidden text-[#525252] lg:inline-flex"
              onClick={toggleSidebar}
              aria-label={
                collapsed ? t("Expand navigation") : t("Collapse navigation")
              }
              title={collapsed ? t("Expand navigation") : t("Collapse navigation")}
            >
              {collapsed ? (
                <PanelLeftOpen size={18} />
              ) : (
                <PanelLeftClose size={18} />
              )}
            </Button>
            <span className="truncate text-sm font-semibold text-[#171717]">
              {pageName}
            </span>
          </div>
          <div className="flex items-center gap-3 lg:hidden">
            <Link href="/workspace" aria-label="Tengeneza CV workspace">
              <BrandMark compact className="app-brand" />
            </Link>
          </div>
          <div className="hidden items-center gap-3 sm:flex">
            <Link
              href="/templates"
              className="text-xs font-medium text-[#525252] transition hover:text-[#004AAD]"
            >
              {t("browseTemplates")}
            </Link>
            <span className="h-4 w-px bg-[#E5E7EB]" />
            <span className="flex items-center gap-2 text-xs font-medium text-[#525252]">
              <i className="h-2 w-2 rounded-full bg-[#004AAD]" />
              {t("Local workspace · Unlimited exports")}
            </span>
          </div>
        </header>
        <main id="main-content" tabIndex={-1}>
          {children}
        </main>
      </div>
      {open && (
        <div
          className="fixed inset-0 z-50 lg:hidden"
          role="dialog"
          aria-modal="true"
          aria-label={t("Main navigation")}
        >
          <button
            onClick={() => setOpen(false)}
            aria-label={t("Close navigation menu")}
            className="absolute inset-0 bg-[#171717]/30"
          />
          <aside className="relative h-full w-[280px] border-r border-[#E5E7EB] bg-white shadow-xl">
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-3 top-3"
              onClick={() => setOpen(false)}
              aria-label={t("Close navigation menu")}
            >
              <X size={20} />
            </Button>
            {nav}
          </aside>
        </div>
      )}
    </div>
  );
}

export function PublicHeader() {
  const location = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const navItems = [
    { href: "/templates", label: "Templates" },
    { href: "/examples", label: "Examples" },
    { href: "/workspace", label: "Workspace" },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-[#E2EEFF] bg-white/85 backdrop-blur-xl transition-all">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-3.5 lg:px-8">
        
        {/* BRAND LOGO */}
        <Link href="/" className="flex items-center gap-2.5 transition hover:opacity-95">
          <BrandMark />
        </Link>

        {/* DESKTOP NAV LINKS */}
        <nav className="hidden items-center gap-1.5 rounded-full border border-[#E2EEFF] bg-[#F4F8FF]/80 p-1 md:flex shadow-xs">
          {navItems.map(item => {
            const isActive = location === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "px-5 py-2 text-xs font-bold transition-all rounded-full",
                  isActive
                    ? "bg-white text-[#004AAD] shadow-sm font-extrabold"
                    : "text-[#48617D] hover:text-[#102A56] hover:bg-white/60"
                )}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* RIGHT CONTROLS & CTA */}
          <div className="hidden items-center gap-3 md:flex">
            <LocaleControl />
            <Link href="/workspace">
              <Button className="rounded-full bg-[#004AAD] px-6 text-xs font-extrabold text-white shadow-md shadow-[#004AAD]/25 hover:bg-[#003B8E] transition-all">
                Open Workspace
              </Button>
            </Link>
          </div>

        {/* MOBILE MENU TOGGLE */}
        <div className="flex items-center gap-2 md:hidden">
          <LocaleControl />
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="text-[#102A56]"
          >
            {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
          </Button>
        </div>

      </div>

      {/* MOBILE DROPDOWN MENU */}
      {mobileMenuOpen && (
        <div className="border-b border-[#E2EEFF] bg-white px-6 py-4 md:hidden shadow-lg space-y-3 animate-in slide-in-from-top-2">
          {navItems.map(item => (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setMobileMenuOpen(false)}
              className={cn(
                "block py-2.5 px-4 text-sm font-bold rounded-lg transition",
                location === item.href
                  ? "bg-[#EAF2FF] text-[#004AAD]"
                  : "text-[#48617D] hover:bg-[#F7FAFF]"
              )}
            >
              {item.label}
            </Link>
          ))}
          <div className="pt-2">
            <Link href="/workspace" onClick={() => setMobileMenuOpen(false)}>
              <Button className="w-full rounded-lg bg-[#004AAD] py-5 text-sm font-bold text-white shadow-md">
                Open Workspace
              </Button>
            </Link>
          </div>
        </div>
      )}
    </header>
  );
}
