'use client';

import { getToken } from "@/lib/auth";
import { redirect } from "next/navigation";

export default function RootPage() {
  const token = getToken();
  redirect(token ? "/workspace" : "/signup");
}
