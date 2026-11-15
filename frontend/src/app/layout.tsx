import type { Metadata } from "next";
import { AuthProvider } from "./auth";
import "./globals.css";

export const metadata: Metadata = {
  title: "Next.js Starter",
  description: "A clean Next.js application starter.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru">
      <body><AuthProvider>{children}</AuthProvider></body>
    </html>
  );
}
