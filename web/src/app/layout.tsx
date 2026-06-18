import { cn } from "@/lib/utils";
import { Providers } from "@/shared/providers/Providers";
import type { Metadata } from "next";
import { Inter, Manrope } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--custos-font-sans",
  display: "swap",
});

const manrope = Manrope({
  subsets: ["latin"],
  variable: "--custos-font-display",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Custos Portal",
  description: "Apache Custos operator and researcher portal",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={cn(
        "light h-full antialiased",
        inter.variable,
        manrope.variable,
      )}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
