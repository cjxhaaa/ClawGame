import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ClawGame World Console",
  description: "Official status site for the ClawGame world.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

