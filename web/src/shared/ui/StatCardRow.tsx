import { cn } from "@/lib/utils";
import type { HTMLAttributes, ReactNode } from "react";

export type StatCardRowCols = 3 | 4 | 5;

export type StatCardRowProps = Omit<HTMLAttributes<HTMLDivElement>, "children"> & {
  children: ReactNode;
  cols?: StatCardRowCols;
};

const COL_CLASS: Record<StatCardRowCols, string> = {
  3: "md:grid-cols-3",
  4: "sm:grid-cols-2 lg:grid-cols-4",
  5: "sm:grid-cols-2 lg:grid-cols-5",
};

export function StatCardRow({ children, cols = 3, className, ...rest }: StatCardRowProps) {
  return (
    <div className={cn("grid grid-cols-1 gap-4", COL_CLASS[cols], className)} {...rest}>
      {children}
    </div>
  );
}
