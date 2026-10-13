/** Nairobi Blue Ledger: a faithful, print-grade folio preview of the canonical CV document. */
import { assetUrl, type CVContent } from "@/lib/api";
import { useLanguage } from "@/contexts/LanguageContext";
const list = (x: any) => (Array.isArray(x) ? x : []);
const roleOf = (x: any) => x.title || x.role || "";
const employerOf = (x: any) => x.employer || x.company || "";
const credentialOf = (x: any) => x.credential || x.qualification || "";
const dateRange = (x: any, locale: "en" | "sw") =>
  [x.start_date, x.end_date || (x.is_current ? (locale === "sw" ? "Sasa" : "Present") : "")]
    .filter(Boolean)
    .join(" — ");
export function CvPreview({
  content,
  title,
  template,
}: {
  content: CVContent;
  title: string;
  template?: { name?: string; family?: string; accent_color_hex?: string };
}) {
  const { locale, t } = useLanguage();
  const p = content.personal || {},
    settings = content.settings || {},
    exp = list(content.experience),
    edu = list(content.education),
    certs = list(content.certifications),
    referees = list(content.referees),
    skills = content.skills || {};
  const contact = [p.email, p.phone, p.location].filter(Boolean);
  const family = template?.family || "ats";
  return (
    <div className="folio-wrap">
      <i className="folio-crop a" />
      <i className="folio-crop b" />
      <article
        className={`cv-folio family-${family}`}
        style={{ ["--folio-accent" as any]: template?.accent_color_hex }}
      >
        <small>
          TENGENEZA CV / {template?.name ? `${template.name} · ` : ""}
          {title} ·{" "}
          {locale === "sw" ? "KISWAHILI" : "ENGLISH"}
        </small>
        <header className={p.photo_url && settings.show_photo !== false ? "has-photo" : ""}>
          {p.photo_url && (
            <img
              className="folio-photo"
              src={assetUrl(p.photo_url)}
              alt={t("Profile")}
            />
          )}
          <h1>{p.full_name || (locale === "sw" ? "Jina Lako" : "Your Name")}</h1>
          {p.headline && <b>{p.headline}</b>}
          <p>{contact.join("  ·  ") || (locale === "sw" ? "barua pepe · simu · mahali" : "email · phone · location")}</p>
          {(p.linkedin_url || p.portfolio_url) && (
            <p className="folio-links">
              {[p.linkedin_url, p.portfolio_url].filter(Boolean).join("  ·  ")}
            </p>
          )}
        </header>
        {content.summary ? (
          <Section title={t("Professional Summary")}>
            <p className="folio-summary">{content.summary}</p>
          </Section>
        ) : (
          <Section title={t("Professional Summary")}>
            <p>{t("Your professional summary will appear here.")}</p>
          </Section>
        )}
        {(list(skills.hard_skills).length > 0 || list(skills.soft_skills).length > 0) && (
          <Section title={t("Key Skills & Expertise")}>
            {list(skills.hard_skills).length > 0 && (
              <p className="folio-chips">
                {list(skills.hard_skills).map((s: string) => (
                  <span key={s}>{s}</span>
                ))}
              </p>
            )}
            {list(skills.soft_skills).length > 0 && (
              <p className="folio-soft-skills" style={{ marginTop: "0.4rem", fontSize: "0.85rem", color: "#475569" }}>
                <strong>{t("Core Strengths: ")}</strong>
                {list(skills.soft_skills).join(" · ")}
              </p>
            )}
          </Section>
        )}
        <Section title={t("Work Experience")}>
          {exp.length ? (
            exp.map((e: any, i: number) => {
              const rawBullets = e.bullets || e.achievements || e.highlights || e.responsibilities || e.description;
              const bulletsList: string[] = Array.isArray(rawBullets)
                ? rawBullets.map((x: any) => (typeof x === "string" ? x : String(x?.item || x?.text || x))).filter(x => x.trim().length > 0)
                : typeof rawBullets === "string" && rawBullets.trim()
                ? rawBullets.split("\n").filter(x => x.trim().length > 0)
                : [];

              return (
                <div className="folio-entry" key={i}>
                  <div className="folio-entry-head">
                    <strong>{roleOf(e) || t("Role title")}</strong>
                    {dateRange(e, locale) && <span>{dateRange(e, locale)}</span>}
                  </div>
                  <p className="accent">
                    {[employerOf(e), e.location].filter(Boolean).join(" · ") ||
                      t("Company")}
                  </p>
                  {bulletsList.length > 0 && (
                    <ul>
                      {bulletsList.map((b: string, j: number) => (
                        <li key={j}>{b.replace(/^[•\-\*\s]+/, "")}</li>
                      ))}
                    </ul>
                  )}
                </div>
              );
            })
          ) : (
            <p>{t("Add your strongest roles and action-led outcomes.")}</p>
          )}
        </Section>
        <div className="folio-grid">
          <Section title={t("Education")}>
            {edu.length ? (
              edu.map((e: any, i: number) => (
                <div className="folio-entry" key={i}>
                  <strong>{e.institution || t("Institution")}</strong>
                  <p>{credentialOf(e) || t("Qualification / Degree")}</p>
                  {dateRange(e, locale) && <small>{dateRange(e, locale)}</small>}
                </div>
              ))
            ) : (
              <p>{t("Education details")}</p>
            )}
          </Section>
          {list(skills.languages).length > 0 && (
            <Section title={t("Languages (comma separated)")}>
              <p>
                {list(skills.languages)
                  .map((l: any) =>
                    typeof l === "string"
                      ? l
                      : [l.language, l.proficiency].filter(Boolean).join(" (") +
                        (l.proficiency ? ")" : "")
                  )
                  .join(" · ")}
              </p>
            </Section>
          )}
        </div>
        {certs.length > 0 && (
          <Section title={t("Certifications")}>
            {certs.map((c: any, i: number) => (
              <div className="folio-entry folio-compact" key={i}>
                <strong>{c.name}</strong>
                <p>
                  {[c.issuer, c.year].filter(Boolean).join(" · ")}
                </p>
              </div>
            ))}
          </Section>
        )}
        {settings.show_referees && referees.length > 0 && (
          <Section title={t("Referees")}>
            <div className="folio-grid">
              {referees.map((r: any, i: number) => (
                <p key={i} className="folio-referee">
                  <strong>{r.name}</strong>
                  <br />
                  {[r.title, r.organization, r.phone, r.email]
                    .filter(Boolean)
                    .join(" · ")}
                </p>
              ))}
            </div>
          </Section>
        )}
      </article>
    </div>
  );
}
function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="cv-section">
      <h2>{title}</h2>
      {children}
    </section>
  );
}
