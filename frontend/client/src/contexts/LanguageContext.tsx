'use client';

import { createContext, useContext, useMemo, useState, useEffect } from "react";

type Locale = "en" | "sw";

const copy: Record<Locale, Record<string, string>> = {
  en: {
    workspace: "Workspace", documents: "Documents", profile: "Profile", plans: "Free & apps",
    templates: "Templates", examples: "Examples", save: "Save changes", create: "New CV",
    signIn: "Sign in", signOut: "Sign out", overview: "Overview", language: "Language",
    english: "English", swahili: "Kiswahili", careerWorkspace: "Career workspace",
    administration: "Administration", cvEditor: "CV editor", browseTemplates: "Browse templates",
    camelConnected: "CamelAccounts connected",
    "Main navigation": "Main navigation", "Local workspace": "Local workspace",
    "Unlimited exports": "Unlimited exports", "Local workspace · Unlimited exports": "Local workspace · Unlimited exports",
    "Skip to content": "Skip to content", "Open navigation menu": "Open navigation menu",
    "Close navigation menu": "Close navigation menu", "Expand navigation": "Expand navigation",
    "Collapse navigation": "Collapse navigation", "Your CV workspace": "Your CV workspace",
    "Create, improve, and keep your career documents in one place.": "Create, improve, and keep your career documents in one place.",
    "Good to see you, {name}": "Good to see you, {name}", "My CVs": "My CVs", "Active documents": "Active documents",
    "Available to use": "Available to use", "Exports": "Exports", "Unlimited": "Unlimited",
    "PDF, DOCX & CSV": "PDF, DOCX & CSV", "Available today": "Available today",
    "Your documents": "Your documents", "Pick up where you left off or start something new.": "Pick up where you left off or start something new.",
    "Your documents could not be loaded": "Your documents could not be loaded", "Try again": "Try again",
    "Open CV": "Open CV", "Duplicate CV": "Duplicate CV", "Archive CV": "Archive CV",
    "Your first CV starts here": "Your first CV starts here", "Choose a professional template, add your experience, then review and export your CV when it is ready.": "Choose a professional template, add your experience, then review and export your CV when it is ready.",
    "Create your first CV": "Create your first CV", "Explore templates": "Explore templates", "Browse all templates": "Browse all templates",
    "Template Design": "Template Design", "Choose your CV design": "Choose your CV design", "Select a design, review your CV, then continue to edit.": "Select a design, review your CV, then continue to edit.",
    "Live preview": "Live preview", "Your CV": "Your CV", "Opening…": "Opening…", "Continue": "Continue",
    "Change template": "Change template", "Save Document": "Save Document", "CV Document Title": "CV Document Title",
    "Personal & Contact": "Personal & Contact", "Profile Photo": "Profile Photo", "Remove photo": "Remove photo",
    "Upload Photo": "Upload Photo", "Uploading...": "Uploading...", "Or paste photo URL...": "Or paste photo URL...",
    "Full Name": "Full Name", "Professional Headline": "Professional Headline", "Email": "Email", "Phone": "Phone",
    "Location": "Location", "LinkedIn URL": "LinkedIn URL", "Portfolio / Website URL": "Portfolio / Website URL",
    "Professional Summary": "Professional Summary", "Work Experience": "Work Experience", "Education": "Education",
    "Skills & Languages": "Skills & Languages", "Certifications": "Certifications", "Referees": "Referees",
    "Role": "Role", "Title / Role": "Title / Role", "Employer": "Employer", "Start Date": "Start Date",
    "End Date": "End Date", "End Date / Graduated": "End Date / Graduated",
    "Key Achievements & Bullet Points (one per line)": "Key Achievements & Bullet Points (one per line)",
    "Add Experience": "Add Experience", "Add Education": "Add Education", "Add Certification": "Add Certification",
    "Add Referee": "Add Referee", "Institution": "Institution", "Qualification / Degree": "Qualification / Degree",
    "Hard Skills (comma separated)": "Hard Skills (comma separated)", "Strengths / Soft Skills (comma separated)": "Strengths / Soft Skills (comma separated)",
    "Languages (comma separated)": "Languages (comma separated)", "Certification": "Certification", "Certification Name": "Certification Name",
    "Issuer": "Issuer", "Year": "Year", "Referee": "Referee", "Organization": "Organization",
    "Preview & export CV": "Preview & export CV", "Back to editing": "Back to editing", "Select color:": "Select color:",
    "Preview an accent color (not saved to the CV)": "Preview an accent color (not saved to the CV)",
    "Preview accent color": "Preview accent color", "Done": "Done", "Pick custom color": "Pick custom color",
    "Explore colors at Rangi Tanzania": "Explore colors at Rangi Tanzania", "More CV tools": "More CV tools",
    "ATS rules check": "ATS rules check", "ATS formatting checks": "ATS formatting checks", "Run check": "Run check",
    "Checking formatting…": "Checking formatting…", "History": "History", "Document History": "Document History",
    "Close": "Close", "Saved version": "Saved version", "Click to restore state": "Click to restore state",
    "No saved versions yet.": "No saved versions yet.", "Download:": "Download:", "Preparing…": "Preparing…",
    "Email CV": "Email CV", "ACCOUNT SETTINGS": "ACCOUNT SETTINGS", "Your professional defaults.": "Your professional defaults.",
    "Set the language and region that should guide each new document.": "Set the language and region that should guide each new document.",
    "Profile photo": "Profile photo", "Replace photo": "Replace photo", "Upload photo": "Upload photo",
    "Full name": "Full name", "Phone number": "Phone number", "Preferred language": "Preferred language",
    "Target region": "Target region", "Saving…": "Saving…", "Save preferences": "Save preferences",
    "Looking for Example Documents?": "Looking for Example Documents?", "Explore 24 live API sample documents with filled career content.": "Explore 24 live API sample documents with filled career content.",
    "View Example Documents": "View Example Documents", "All templates": "All templates", "ALL UNLOCKED · FREE EXPORTS": "ALL UNLOCKED · FREE EXPORTS",
    "Preview": "Preview", "Use design": "Use design", "Rendered Template Preview": "Rendered Template Preview",
    "PDF/DOCX RENDER MATCH": "PDF/DOCX RENDER MATCH", "Start clean document": "Start clean document",
    "Use template with sample data": "Use template with sample data", "Template not found": "Template not found",
    "← Back to all templates": "← Back to all templates", "Back to Templates": "Back to Templates",
    "ATS-Optimized Formatting": "ATS-Optimized Formatting", "Photo & portrait layout": "Photo & portrait layout",
    "Clean text layout": "Clean text layout", "Export clean PDF & DOCX": "Export clean PDF & DOCX",
    "Start with {name}": "Start with {name}", "EXAMPLE DOCUMENTS FROM THE API": "EXAMPLE DOCUMENTS FROM THE API",
    "See finished CVs built on these designs": "See finished CVs built on these designs", "Back to example selection": "Back to example selection",
    "Search examples by role...": "Search examples by role...", "Example": "Example", "No examples match your search.": "No examples match your search.",
    "Accent Palette": "Accent Palette", "Rendered as {name}": "Rendered as {name}",
    "Click color swatch to customize accent theme live.": "Click color swatch to customize accent theme live.", "Clone this example": "Clone this example",
    "Key Skills & Expertise": "Key Skills & Expertise", "Core Strengths: ": "Core Strengths: ", "Your professional summary will appear here.": "Your professional summary will appear here.", "Add your strongest roles and action-led outcomes.": "Add your strongest roles and action-led outcomes.", "Role title": "Role title", "Company": "Company", "Present": "Present", "Creating CV…": "Creating CV…", "Family": "Family", "photo-ready": "photo-ready", "Template": "Template", "Accent color: {color}": "Accent color: {color}", "Click to select or type accent color": "Click to select or type accent color", "Changes saved.": "Changes saved.", "ATS score updated.": "ATS score updated.", "Add your headline to help employers understand your experience.": "Add your headline to help employers understand your experience.",
  },
  sw: {
    workspace: "Eneo la kazi", documents: "Nyaraka", profile: "Wasifu", plans: "Bure na programu",
    templates: "Violezo", examples: "Mifano", save: "Hifadhi mabadiliko", create: "CV mpya",
    signIn: "Ingia", signOut: "Toka", overview: "Muhtasari", language: "Lugha",
    english: "Kiingereza", swahili: "Kiswahili", careerWorkspace: "Eneo la kazi za taaluma",
    administration: "Usimamizi", cvEditor: "Kihariri cha CV", browseTemplates: "Vinjari violezo",
    camelConnected: "CamelAccounts imeunganishwa",
    "Main navigation": "Menyu kuu", "Local workspace": "Eneo la kazi la ndani",
    "Unlimited exports": "Uhamishaji bila kikomo", "Local workspace · Unlimited exports": "Eneo la kazi la ndani · Uhamishaji bila kikomo",
    "Skip to content": "Ruka hadi kwenye maudhui", "Open navigation menu": "Fungua menyu ya urambazaji",
    "Close navigation menu": "Funga menyu ya urambazaji", "Expand navigation": "Panua urambazaji",
    "Collapse navigation": "Kunja urambazaji", "Your CV workspace": "Eneo lako la CV",
    "Create, improve, and keep your career documents in one place.": "Unda, boresha na hifadhi nyaraka zako za taaluma mahali pamoja.",
    "Good to see you, {name}": "Tunafurahi kukuona, {name}", "My CVs": "CV zangu", "Active documents": "Nyaraka zinazotumika",
    "Available to use": "Vinavyopatikana", "Exports": "Uhamishaji", "Unlimited": "Bila kikomo",
    "PDF, DOCX & CSV": "PDF, DOCX na CSV", "Available today": "Vinavyopatikana leo",
    "Your documents": "Nyaraka zako", "Pick up where you left off or start something new.": "Endelea ulipoishia au anza jambo jipya.",
    "Your documents could not be loaded": "Nyaraka zako hazikuweza kupakiwa", "Try again": "Jaribu tena",
    "Open CV": "Fungua CV", "Duplicate CV": "Nakili CV", "Archive CV": "Hifadhi CV kwenye kumbukumbu",
    "Your first CV starts here": "CV yako ya kwanza inaanzia hapa", "Choose a professional template, add your experience, then review and export your CV when it is ready.": "Chagua kiolezo cha kitaalamu, ongeza uzoefu wako, kisha kagua na upakue CV yako ikiwa tayari.",
    "Create your first CV": "Unda CV yako ya kwanza", "Explore templates": "Chunguza violezo", "Browse all templates": "Vinjari violezo vyote",
    "Template Design": "Muundo wa kiolezo", "Choose your CV design": "Chagua muundo wa CV yako", "Select a design, review your CV, then continue to edit.": "Chagua muundo, kagua CV yako, kisha endelea kuhariri.",
    "Live preview": "Onyesho la moja kwa moja", "Your CV": "CV yako", "Opening…": "Inafunguliwa…", "Continue": "Endelea",
    "Change template": "Badilisha kiolezo", "Save Document": "Hifadhi hati", "CV Document Title": "Kichwa cha hati ya CV",
    "Personal & Contact": "Taarifa binafsi na mawasiliano", "Profile Photo": "Picha ya wasifu", "Remove photo": "Ondoa picha",
    "Upload Photo": "Pakia picha", "Uploading...": "Inapakia...", "Or paste photo URL...": "Au bandika anwani ya picha...",
    "Full Name": "Jina kamili", "Professional Headline": "Kichwa cha kitaaluma", "Email": "Barua pepe", "Phone": "Simu",
    "Location": "Mahali", "LinkedIn URL": "Anwani ya LinkedIn", "Portfolio / Website URL": "Anwani ya jalada / tovuti",
    "Professional Summary": "Muhtasari wa kitaaluma", "Work Experience": "Uzoefu wa kazi", "Education": "Elimu",
    "Skills & Languages": "Ujuzi na lugha", "Certifications": "Vyeti", "Referees": "Wadhamini",
    "Role": "Wadhifa", "Title / Role": "Cheo / Wadhifa", "Employer": "Mwajiri", "Start Date": "Tarehe ya kuanza",
    "End Date": "Tarehe ya mwisho", "End Date / Graduated": "Tarehe ya mwisho / Kuhitimu",
    "Key Achievements & Bullet Points (one per line)": "Mafanikio muhimu (jambo moja kwa kila mstari)",
    "Add Experience": "Ongeza uzoefu", "Add Education": "Ongeza elimu", "Add Certification": "Ongeza cheti",
    "Add Referee": "Ongeza mdhamini", "Institution": "Taasisi", "Qualification / Degree": "Sifa / Shahada",
    "Hard Skills (comma separated)": "Ujuzi wa kitaalamu (tenganisha kwa koma)", "Strengths / Soft Skills (comma separated)": "Uwezo binafsi (tenganisha kwa koma)",
    "Languages (comma separated)": "Lugha (tenganisha kwa koma)", "Certification": "Cheti", "Certification Name": "Jina la cheti",
    "Issuer": "Mtoa cheti", "Year": "Mwaka", "Referee": "Mdhamini", "Organization": "Shirika",
    "Preview & export CV": "Hakiki na upakue CV", "Back to editing": "Rudi kuhariri", "Select color:": "Chagua rangi:",
    "Preview an accent color (not saved to the CV)": "Hakiki rangi ya msisitizo (haitahifadhiwa kwenye CV)",
    "Preview accent color": "Hakiki rangi ya msisitizo", "Done": "Imekamilika", "Pick custom color": "Chagua rangi maalum",
    "Explore colors at Rangi Tanzania": "Chunguza rangi kwenye Rangi Tanzania", "More CV tools": "Zana zaidi za CV",
    "ATS rules check": "Kagua vigezo vya ATS", "ATS formatting checks": "Ukaguzi wa muundo wa ATS", "Run check": "Anza ukaguzi",
    "Checking formatting…": "Inakagua muundo…", "History": "Historia", "Document History": "Historia ya hati",
    "Close": "Funga", "Saved version": "Toleo lililohifadhiwa", "Click to restore state": "Bofya kurejesha hali hii",
    "No saved versions yet.": "Hakuna matoleo yaliyohifadhiwa bado.", "Download:": "Pakua:", "Preparing…": "Inaandaa…",
    "Email CV": "Tuma CV kwa barua pepe", "ACCOUNT SETTINGS": "MIPANGILIO YA AKAUNTI", "Your professional defaults.": "Mipangilio yako ya kitaaluma.",
    "Set the language and region that should guide each new document.": "Weka lugha na eneo vitakavyotumika kwa kila hati mpya.",
    "Profile photo": "Picha ya wasifu", "Replace photo": "Badilisha picha", "Upload photo": "Pakia picha",
    "Full name": "Jina kamili", "Phone number": "Nambari ya simu", "Preferred language": "Lugha unayopendelea",
    "Target region": "Eneo lengwa", "Saving…": "Inahifadhi…", "Save preferences": "Hifadhi mapendeleo",
    "Looking for Example Documents?": "Unatafuta nyaraka za mfano?", "Explore 24 live API sample documents with filled career content.": "Chunguza nyaraka 24 za mfano zenye maudhui ya taaluma.",
    "View Example Documents": "Tazama nyaraka za mfano", "All templates": "Violezo vyote", "ALL UNLOCKED · FREE EXPORTS": "VYOTE VIKO WAZI · UHAMISHAJI BILA MALIPO",
    "Preview": "Hakiki", "Use design": "Tumia muundo", "Rendered Template Preview": "Onyesho la kiolezo", "PDF/DOCX RENDER MATCH": "MUONEKANO WA PDF/DOCX UNAFANANA",
    "Start clean document": "Anza hati tupu", "Use template with sample data": "Tumia kiolezo chenye data ya mfano", "Template not found": "Kiolezo hakikupatikana",
    "← Back to all templates": "← Rudi kwenye violezo vyote", "Back to Templates": "Rudi kwenye violezo",
    "ATS-Optimized Formatting": "Muundo ulioboreshwa kwa ATS", "Photo & portrait layout": "Muundo wenye picha", "Clean text layout": "Muundo safi wa maandishi",
    "Export clean PDF & DOCX": "Hamisha PDF na DOCX safi", "Start with {name}": "Anza na {name}", "EXAMPLE DOCUMENTS FROM THE API": "NYARAKA ZA MFANO KUTOKA API",
    "See finished CVs built on these designs": "Tazama CV zilizokamilika kwa miundo hii", "Back to example selection": "Rudi kwenye orodha ya mifano",
    "Search examples by role...": "Tafuta mifano kwa wadhifa...", "Example": "Mfano", "No examples match your search.": "Hakuna mifano inayolingana na utafutaji wako.",
    "Accent Palette": "Rangi za msisitizo", "Rendered as {name}": "Imeonyeshwa kama {name}",
    "Click color swatch to customize accent theme live.": "Bofya rangi ili kubadilisha muonekano mara moja.", "Clone this example": "Nakili mfano huu",
    "Profile": "Wasifu", "Education details": "Maelezo ya elimu", "Key Skills & Expertise": "Ujuzi Muhimu na Utaalamu", "Core Strengths: ": "Uwezo Mkuu: ", "Your professional summary will appear here.": "Muhtasari wako wa kitaaluma utaonekana hapa.", "Add your strongest roles and action-led outcomes.": "Ongeza nyadhifa zako muhimu na mafanikio yaliyotokana na hatua.", "Role title": "Cheo cha wadhifa", "Company": "Kampuni", "Present": "Sasa", "Creating CV…": "Inaunda CV…", "Family": "Familia", "photo-ready": "tayari kwa picha", "Template": "Kiolezo", "Accent color: {color}": "Rangi ya msisitizo: {color}", "Click to select or type accent color": "Bofya kuchagua au kuandika rangi ya msisitizo", "Changes saved.": "Mabadiliko yamehifadhiwa.", "ATS score updated.": "Alama ya ATS imesasishwa.", "Add your headline to help employers understand your experience.": "Ongeza kichwa kifupi kinachowaeleza waajiri kuhusu uzoefu wako.",
  },
};

type CopyKey = string;
const C = createContext<{
  locale: Locale;
  setLocale: (x: Locale) => void;
  hydrateLocale: (x?: string) => void;
  t: (x: CopyKey) => string;
}>({
  locale: "en",
  setLocale: () => undefined,
  hydrateLocale: () => undefined,
  t: x => copy.en[x] || x,
});

export function LanguageProvider({ children }: { children: React.ReactNode }) {
  const [locale, set] = useState<Locale>("en");

  useEffect(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("tengenezacv_locale") as Locale;
      if (saved === "en" || saved === "sw") set(saved);
    }
  }, []);

  const value = useMemo(
    () => ({
      locale,
      hydrateLocale: (preferred?: string) => {
        const next = preferred === "sw" || preferred === "en" ? preferred : undefined;
        if (next) {
          const explicit = typeof window !== "undefined" ? localStorage.getItem("tengenezacv_locale") : null;
          if (explicit === "en" || explicit === "sw") return;
          set(next);
          if (typeof window !== "undefined") localStorage.setItem("tengenezacv_locale", next);
        }
      },
      setLocale: (next: Locale) => {
        if (typeof window !== "undefined") localStorage.setItem("tengenezacv_locale", next);
        set(next);
      },
      t: (key: CopyKey) => copy[locale]?.[key] || copy.en[key] || key,
    }),
    [locale]
  );

  return <C.Provider value={value}>{children}</C.Provider>;
}

export const useLanguage = () => useContext(C);
