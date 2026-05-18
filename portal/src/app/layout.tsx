import type { Metadata } from "next";
import type React from "react";
import "./globals.css";
import { PortalLayout } from "./layout/PortalLayout";

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
    <html lang="en">
      <body>
        <PortalLayout>{children}</PortalLayout>
      </body>
    </html>
  );
}
