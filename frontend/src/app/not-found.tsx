'use client';

import { BrandMark } from "@/components/BrandMark";
import { Button } from "@/components/ui/button";
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-[#F7FAFF] px-5 text-center">
      <BrandMark />
      <p className="ledger-label mt-14">
        <span /> MISSING PAGE
      </p>
      <h1 className="mt-4 font-serif text-6xl">404</h1>
      <p className="mt-4 max-w-md text-[#48617D]">
        This route does not lead to a document workspace. Return to a clear next
        step.
      </p>
        <Link href="/workspace">
          <Button className="mt-8 bg-[#004AAD] text-white">Go to workspace</Button>
        </Link>
    </div>
  );
}
