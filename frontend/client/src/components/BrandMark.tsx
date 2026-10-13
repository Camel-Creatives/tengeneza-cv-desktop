import { cn } from "@/lib/utils";

import { useState } from "react";

function Mark({ className }: { className?: string }) {
  const [error, setError] = useState(false);

  if (error) {
    return (
      <div className={cn("flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#004AAD] text-xs font-black text-white shadow-sm ring-1 ring-[#004AAD]/20", className)}>
        TCV
      </div>
    );
  }

  return (
    <img
      src="/6.png"
      alt="Tengeneza CV"
      onError={() => setError(true)}
      className={cn("h-9 w-9 shrink-0 rounded-full object-cover shadow-sm ring-1 ring-[#004AAD]/20", className)}
    />
  );
}

export function BrandMark({ compact = false, className }: { compact?: boolean; className?: string }) {
  return (
    <div className={cn("flex items-center gap-3", className)}>
      <Mark />
      <span className={cn("font-black tracking-[.18em] text-[#102A56]", compact ? "text-sm" : "text-base")}>
        TENGENEZA<span className="ml-1 text-[#004AAD]">CV</span>
      </span>
    </div>
  );
}

