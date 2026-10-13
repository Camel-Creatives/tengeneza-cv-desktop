import "@/index.css";
import { LanguageProvider } from "@/contexts/LanguageContext";
import { Toaster } from "sonner";
import type { Metadata } from "next";
import { Bricolage_Grotesque, Outfit } from "next/font/google";
import Script from "next/script";

const outfit = Outfit({
  subsets: ["latin"],
  variable: "--font-outfit",
  display: "swap",
});

const bricolage = Bricolage_Grotesque({
  subsets: ["latin"],
  variable: "--font-bricolage",
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL("https://app.tengenezacv.online"),
  title: "Tengeneza CV — Professional Career Workspace",
  description:
    "Create, edit, and export ATS-friendly professional CVs in PDF and DOCX formats.",
  icons: {
    icon: [{ url: "/favicon.png", type: "image/png", sizes: "64x64" }],
    shortcut: [{ url: "/favicon.png", type: "image/png" }],
    apple: [{ url: "/apple-touch-icon.png", type: "image/png", sizes: "180x180" }],
  },
  openGraph: {
    type: "website",
    siteName: "Tengeneza CV",
    title: "Tengeneza CV — Professional Career Workspace",
    description:
      "Create, edit, and export ATS-friendly professional CVs in PDF and DOCX formats.",
    url: "https://app.tengenezacv.online",
    images: [{
      url: "/opengraph-image.png",
      width: 1200,
      height: 630,
      alt: "Tengeneza CV robot logo and wordmark",
    }],
  },
  twitter: {
    card: "summary_large_image",
    title: "Tengeneza CV — Professional Career Workspace",
    description:
      "Create, edit, and export ATS-friendly professional CVs in PDF and DOCX formats.",
    images: [{ url: "/opengraph-image.png", alt: "Tengeneza CV robot logo and wordmark" }],
  },
  applicationName: "Tengeneza CV",
  appleWebApp: { capable: true, title: "Tengeneza CV", statusBarStyle: "default" },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${outfit.variable} ${bricolage.variable}`}>
      <body className="bg-background text-foreground antialiased">
        <LanguageProvider>
          {children}
          <Toaster position="top-right" richColors />
        </LanguageProvider>
      </body>
      <Script
        src="https://analytics.camelcreatives.com/tracker.js"
        data-site="tengeneza-cv"
        strategy="afterInteractive"
      />
    </html>
  );
}
