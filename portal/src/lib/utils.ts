// Shared `cn` helper that combines clsx's conditional class composition
// with tailwind-merge's last-wins conflict resolution. Used pervasively by
// the shadcn/ui components and any component that accepts a `className`
// override.
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
