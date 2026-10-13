import { cn } from "@/lib/utils";

type LoadingKind = "workspace" | "editor" | "cards" | "profile";

export function AppLoading({ kind }: { kind: LoadingKind }) {
  const cardCount = kind === "cards" ? 6 : kind === "workspace" ? 3 : 2;

  return (
    <div
      className="min-h-screen bg-[#F7F7F7]"
      role="status"
      aria-live="polite"
      aria-label="Loading Tengeneza CV"
    >
      <span className="sr-only">
        Loading {kind === "cards" ? "your documents" : kind}…
      </span>
      <aside className="fixed inset-y-0 left-0 hidden w-[252px] border-r border-[#E5E7EB] bg-white p-6 lg:block">
        <div className="h-8 w-36 animate-pulse rounded bg-[#F3F7FD]" />
        <div className="mt-12 h-3 w-28 animate-pulse rounded bg-[#F5F5F5]" />
        <div className="mt-4 space-y-3">
          {[0, 1, 2].map(item => (
            <div
              key={item}
              className="h-10 animate-pulse rounded-md bg-[#F7F7F7]"
            />
          ))}
        </div>
        <div className="absolute inset-x-6 bottom-6 h-px bg-[#E5E7EB]" />
      </aside>
      <div className="lg:pl-[252px]">
        <header className="flex h-16 items-center justify-between border-b border-[#E5E7EB] bg-white px-5">
          <div className="h-4 w-24 animate-pulse rounded bg-[#F1F1F1]" />
          <div className="h-4 w-36 animate-pulse rounded bg-[#F1F1F1]" />
        </header>
        <main className="mx-auto min-h-[calc(100vh-64px)] max-w-[1440px] px-4 py-7 sm:px-6 lg:px-10">
          <div className="flex items-end justify-between border-b border-[#E5E7EB] pb-6">
            <div className="space-y-3">
              <div className="h-7 w-56 max-w-[65vw] animate-pulse rounded bg-[#E5E7EB]" />
              <div className="h-4 w-80 max-w-[75vw] animate-pulse rounded bg-[#F0F0F0]" />
            </div>
            <div className="hidden h-10 w-32 animate-pulse rounded-md bg-[#CFDFFF] sm:block" />
          </div>
          {kind === "editor" ? (
            <div className="mt-6 grid gap-5 xl:grid-cols-[minmax(360px,.8fr)_minmax(520px,1.2fr)]">
              <div className="space-y-4 rounded-lg border border-[#E5E7EB] bg-white p-5">
                {[0, 1, 2, 3, 4, 5].map(item => (
                  <div
                    key={item}
                    className={cn(
                      "h-11 animate-pulse rounded-md bg-[#F5F5F5]",
                      item === 2 && "h-24"
                    )}
                  />
                ))}
              </div>
              <div className="flex justify-center rounded-lg border border-[#E5E7EB] bg-[#F1F1F1] p-6">
                <div className="min-h-[640px] w-full max-w-[480px] animate-pulse border border-[#E5E7EB] bg-white p-8">
                  <div className="h-2 w-24 rounded bg-[#CFDFFF]" />
                  <div className="mt-8 h-6 w-48 rounded bg-[#E5E7EB]" />
                  <div className="mt-3 h-3 w-32 rounded bg-[#F0F0F0]" />
                  <div className="mt-8 space-y-3">
                    {[0, 1, 2, 3, 4, 5, 6, 7].map(item => (
                      <div key={item} className="h-2 rounded bg-[#F1F1F1]" />
                    ))}
                  </div>
                </div>
              </div>
            </div>
          ) : kind === "profile" ? (
            <div className="mt-6 grid gap-4 lg:grid-cols-[.7fr_1.3fr]">
              <div className="h-56 animate-pulse rounded-lg border border-[#E5E7EB] bg-white" />
              <div className="h-[420px] animate-pulse rounded-lg border border-[#E5E7EB] bg-white" />
            </div>
          ) : (
            <div
              className={cn(
                "mt-6 grid gap-4 md:grid-cols-2",
                kind === "cards" && "xl:grid-cols-3"
              )}
            >
              {Array.from({ length: cardCount }, (_, item) => (
                <div
                  key={item}
                  className="overflow-hidden rounded-lg border border-[#E5E7EB] bg-white"
                >
                  <div className="h-[150px] animate-pulse bg-[#F3F7FD]" />
                  <div className="space-y-3 p-4">
                    <div className="h-4 w-28 animate-pulse rounded bg-[#F0F0F0]" />
                    <div className="h-5 w-2/3 animate-pulse rounded bg-[#E5E7EB]" />
                    <div className="h-4 w-full animate-pulse rounded bg-[#F0F0F0]" />
                    <div className="h-8 w-full animate-pulse rounded bg-[#F8F8F8]" />
                  </div>
                </div>
              ))}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
