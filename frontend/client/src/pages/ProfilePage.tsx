'use client';

/** Nairobi Blue Ledger: account choices give language, region, and photo preferences a clear home. */
import { AppShell } from "@/components/AppShell";
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
import { api, assetUrl, User } from "@/lib/api";
import { useLanguage } from "@/contexts/LanguageContext";
import { getToken, useAuthGuard } from "@/lib/auth";
import { Camera, Loader2, Save } from "lucide-react";
import { ChangeEvent, useEffect, useState } from "react";
import { toast } from "sonner";

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <Label className="text-xs font-extrabold text-[#525252]">{label}</Label>
      {children}
    </div>
  );
}

export function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const { hydrateLocale, t } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const token = getToken();
  useAuthGuard();

  useEffect(() => {
    api
      .me(token)
      .then(profile => { setUser(profile); hydrateLocale(profile.preferred_language); })
      .catch(e =>
        toast.error(e instanceof Error ? e.message : "Unable to load profile.")
      )
      .finally(() => setLoading(false));
  }, []);

  async function save() {
    if (!user) return;
    setSaving(true);
    try {
      const x = await api.updateMe(token, user);
      setUser(x);
      hydrateLocale(x.preferred_language);
      toast.success("Profile preferences saved.");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Could not save profile.");
    } finally {
      setSaving(false);
    }
  }

  async function photo(e: ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0];
    e.target.value = "";
    if (!f || !user) return;
    if (!["image/jpeg", "image/png", "image/webp"].includes(f.type)) {
      toast.error("Choose a JPEG, PNG, or WebP image.");
      return;
    }
    if (f.size > 5 * 1024 * 1024) {
      toast.error("Choose an image smaller than 5 MB.");
      return;
    }
    try {
      const r = await api.uploadPhoto(token, f);
      const nextUser = { ...user, profile_photo_url: r.profile_photo_url };
      setUser(nextUser);
      toast.success("Profile photo uploaded.");
    } catch (x) {
      toast.error(x instanceof Error ? x.message : "Photo upload failed.");
    }
  }

  return (
    <AppShell>
      <div className="workspace-canvas max-w-4xl px-5 py-8 lg:px-10 lg:py-11">
        {loading || !user ? (
          <div className="flex min-h-[50vh] items-center justify-center">
            <Loader2 className="animate-spin text-[#004AAD]" />
          </div>
        ) : (
          <>
            <p className="ledger-label">
              <span /> {t("ACCOUNT SETTINGS")}
            </p>
            <h1 className="mt-3 text-4xl font-extrabold text-[#171717]">
              {t("Your professional defaults.")}
            </h1>
            <p className="mt-3 max-w-xl leading-7 text-[#525252]">
              {t("Set the language and region that should guide each new document.")}
            </p>
            <div className="mt-10 grid gap-5 lg:grid-cols-[.75fr_1.25fr]">
              <section className="rounded-lg border border-[#E5E7EB] bg-white p-6">
                <div className="relative mx-auto flex h-28 w-28 items-center justify-center overflow-hidden rounded-full bg-[#EAF2FF] text-3xl font-extrabold text-[#004AAD]">
                  {user.profile_photo_url ? (
                    <img
                      className="h-full w-full object-cover"
                      src={assetUrl(user.profile_photo_url)}
                      alt={t("Profile")}
                    />
                  ) : (
                    user.full_name.slice(0, 1)
                  )}
                  <label
                    htmlFor="profile-photo-input"
                    className="absolute bottom-0 right-0 flex h-9 w-9 cursor-pointer items-center justify-center rounded-full border-2 border-white bg-[#004AAD] text-white"
                    title={t("Upload photo")}
                  >
                    <Camera size={15} />
                  </label>
                </div>
                <h2 className="mt-5 text-center font-extrabold">
                  {t("Profile photo")}
                </h2>
                <p className="mt-2 text-center text-sm leading-6 text-[#525252]">
                  JPEG, PNG, or WebP up to 5 MB. Use it in photo-supported
                  templates.
                </p>
                <label
                  htmlFor="profile-photo-input"
                  className="mx-auto mt-4 flex w-fit cursor-pointer items-center gap-2 rounded-md bg-[#004AAD] px-4 py-2 text-sm font-bold text-white hover:bg-[#003B8E]"
                >
                  <Camera size={15} />
                  {user.profile_photo_url ? t("Replace photo") : t("Upload photo")}
                </label>
                <input
                  id="profile-photo-input"
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  className="sr-only"
                  onChange={photo}
                />
              </section>
              <section className="rounded-lg border border-[#E5E7EB] bg-white p-6">
                <div className="grid gap-5 sm:grid-cols-2">
                  <Field label={t("Full name")}>
                    <Input
                      value={user.full_name || ""}
                      onChange={e =>
                        setUser({ ...user, full_name: e.target.value })
                      }
                    />
                  </Field>
                  <Field label={t("Email")}>
                    <Input disabled value={user.email || ""} />
                  </Field>
                  <Field label={t("Phone number")}>
                    <Input disabled value={user.phone || ""} />
                  </Field>
                  <Field label={t("Preferred language")}>
                    <Select
                      value={user.preferred_language || "en"}
                      onValueChange={preferred_language =>
                        setUser({ ...user, preferred_language })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="en">English</SelectItem>
                        <SelectItem value="sw">Kiswahili</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                  <Field label={t("Target region")}>
                    <Select
                      value={user.target_region || "eac"}
                      onValueChange={target_region =>
                        setUser({ ...user, target_region })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="eac">East African Community</SelectItem>
                        <SelectItem value="gulf">Gulf region</SelectItem>
                        <SelectItem value="uk">United Kingdom</SelectItem>
                        <SelectItem value="eu">European Union</SelectItem>
                        <SelectItem value="us">United States</SelectItem>
                        <SelectItem value="remote_global">Remote / Global</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                </div>
                <Button
                  onClick={save}
                  disabled={saving}
                  className="mt-7 bg-[#004AAD] text-white hover:bg-[#003B8E]"
                >
                  <Save className="mr-2" size={16} />
                  {saving ? t("Saving…") : t("Save preferences")}
                </Button>
              </section>
            </div>
          </>
        )}
      </div>
    </AppShell>
  );
}
