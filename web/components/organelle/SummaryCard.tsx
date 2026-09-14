import type { ReactNode } from "react";

type SummaryCardProps = {
  title: string;
  children: ReactNode;
};

export function SummaryCard({ title, children }: SummaryCardProps) {
  return (
    <div className="rounded-2xl bg-white p-4 shadow-sm">
      <p className="mb-2 font-medium text-zinc-900">{title}</p>
      {children}
    </div>
  );
}
