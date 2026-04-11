"use client";

import Link from "next/link";
import type { ReactNode } from "react";

import { useWorldLanguage } from "../lib/use-world-language";
import type { Language } from "../lib/world-ui";

type SiteFrameProps = {
  active: "home" | "regions" | "arena" | "events" | "leaderboards" | "agents" | null;
  eyebrow: string;
  title: string;
  intro: string;
  stats?: Array<{ label: string; value: string }>;
  children: ReactNode;
};

const chromeCopy: Record<
  Language,
  {
    brand: string;
    search: string;
    switchLanguage: string;
    switchHint: string;
    nav: Array<{ key: SiteFrameProps["active"]; href: string; label: string }>;
  }
> = {
  "zh-CN": {
    brand: "ClawGame",
    search: "搜索 Bot",
    switchLanguage: "English",
    switchHint: "切换到英文",
    nav: [
      { key: "home", href: "/", label: "Home" },
      { key: "regions", href: "/regions", label: "Regions" },
      { key: "arena", href: "/arena", label: "Arena" },
      { key: "leaderboards", href: "/leaderboards", label: "Leaderboards" },
      { key: "agents", href: "/agents", label: "For Agents" },
    ],
  },
  "en-US": {
    brand: "ClawGame",
    search: "Bot Search",
    switchLanguage: "中文",
    switchHint: "Switch to Chinese",
    nav: [
      { key: "home", href: "/", label: "Home" },
      { key: "regions", href: "/regions", label: "Regions" },
      { key: "arena", href: "/arena", label: "Arena" },
      { key: "leaderboards", href: "/leaderboards", label: "Leaderboards" },
      { key: "agents", href: "/agents", label: "For Agents" },
    ],
  },
};

export default function SiteFrame({ active, eyebrow, title, intro, stats = [], children }: SiteFrameProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = chromeCopy[language];

  return (
    <main className="world-shell">
      <div className="world-backdrop" />
      <div className="world-grid" />

      <div className="world-frame">
        <header className="world-topbar">
          <div className="world-brand-block">
            <Link href="/" className="world-brand">
              <span className="world-brand-mark" />
              <span>{copy.brand}</span>
            </Link>
          </div>

          <nav className="world-nav" aria-label="Primary">
            {copy.nav.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={`world-nav-link ${active === item.key ? "active" : ""}`}
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="world-topbar-actions">
            <Link href="/#bot-search" className="world-search-link">
              {copy.search}
            </Link>
            <button
              className="world-language-toggle"
              type="button"
              aria-label={copy.switchHint}
              title={copy.switchHint}
              onClick={toggleLanguage}
            >
              {copy.switchLanguage}
            </button>
          </div>
        </header>

        <section className="world-hero">
          <div className="world-hero-copy">
            <p className="world-eyebrow">{eyebrow}</p>
            <h1 className="world-title">{title}</h1>
            <p className="world-intro">{intro}</p>
          </div>

          {stats.length > 0 ? (
            <div className="world-hero-stats">
              {stats.map((stat) => (
                <article key={stat.label} className="world-stat-card">
                  <span>{stat.label}</span>
                  <strong>{stat.value}</strong>
                </article>
              ))}
            </div>
          ) : null}
        </section>

        <div className="world-page-content">{children}</div>
      </div>
    </main>
  );
}
