import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import "./globals.css";

const mono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "WASS | Web App Security Scanner",
  description: "A brutalist, SOC-grade frontend for the WASS scanner.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${mono.className} bg-[#0a0a0a] text-[#e5e5e5] antialiased selection:bg-red-700 selection:text-[#e5e5e5]`}
      >
        {children}
      </body>
    </html>
  );
}
