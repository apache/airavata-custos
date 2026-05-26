// Root App Router layout. Loads global styles, registers the Geist font
// CSS variable, wraps the tree in the NextAuth SessionProvider, and
// renders the persistent PortalLayout chrome around each route.
import type { Metadata } from "next";
import type React from "react";
import "./globals.css";
import { PortalLayout } from "./layout/PortalLayout";
import { SessionProviderWrapper } from "./components/SessionProviderWrapper";
import { Geist } from "next/font/google";
import { cn } from "@/lib/utils";

const geist = Geist({ subsets: ["latin"], variable: "--font-sans" });

export const metadata: Metadata = {
  title: "Custos Portal",
  description: "Generic SSH certificate management portal",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={cn("font-sans", geist.variable)}>
      <body>
        <SessionProviderWrapper>
          <PortalLayout>{children}</PortalLayout>
        </SessionProviderWrapper>
      </body>
    </html>
  );
}
