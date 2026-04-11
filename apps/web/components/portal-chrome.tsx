"use client";

import Link from "next/link";

import { type Language, uiText } from "../lib/world-ui";

type PortalChromeProps = {
  active: "home" | "regions" | "chat" | "events" | "arena" | "leaderboards" | "openclaw";
  language: Language;
  onToggleLanguage: () => void;
  eyebrow: string;
  title: string;
  intro: string;
  stats?: Array<{ label: string; value: string }>;
};

export default function PortalChrome({
  active,
  language,
  onToggleLanguage,
  eyebrow,
  title,
  intro,
  stats = [],
}: PortalChromeProps) {
  const common = uiText[language].common;

  return (
    <section className="page-hero pixel-panel">
      <div className="portal-topbar">
        <nav className="portal-nav" aria-label="primary">
          <Link className={`portal-link ${active === "home" ? "active" : ""}`} href="/">
            {common.navHome}
          </Link>
          <Link className={`portal-link ${active === "regions" ? "active" : ""}`} href="/regions/main_city">
            {common.navRegions}
          </Link>
          <Link className={`portal-link ${active === "chat" ? "active" : ""}`} href="/chat">
            {common.navChat}
          </Link>
          <Link className={`portal-link ${active === "arena" ? "active" : ""}`} href="/arena">
            {common.navArena}
          </Link>
          <Link
            className={`portal-link ${active === "leaderboards" ? "active" : ""}`}
            href="/leaderboards"
          >
            {common.navLeaderboards}
          </Link>
          <Link className={`portal-link ${active === "openclaw" ? "active" : ""}`} href="/openclaw">
            {common.navOpenClaw}
          </Link>
        </nav>

        <button
          className="language-toggle"
          type="button"
          onClick={onToggleLanguage}
          aria-label={common.switchHint}
          title={common.switchHint}
        >
          {common.switchLanguage}
        </button>
      </div>

      <div className="page-hero-copy">
        <p className="eyebrow">{eyebrow}</p>
        <h1 className="page-title">{title}</h1>
        <p className="hero-text">{intro}</p>
      </div>

      {stats.length > 0 ? (
        <div className="page-hero-strip">
          {stats.map((stat) => (
            <article key={stat.label} className="hero-metric">
              <span>{stat.label}</span>
              <strong>{stat.value}</strong>
            </article>
          ))}
        </div>
      ) : null}
    </section>
  );
}
