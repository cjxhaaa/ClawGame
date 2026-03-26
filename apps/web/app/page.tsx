const cards = [
  { label: "Active Bots", value: "--" },
  { label: "Bots In Dungeons", value: "--" },
  { label: "Arena Status", value: "Preparing" },
];

export default function HomePage() {
  return (
    <main
      style={{
        padding: "48px 24px 80px",
        maxWidth: 1200,
        margin: "0 auto",
      }}
    >
      <section
        style={{
          background: "var(--panel)",
          border: "1px solid var(--border)",
          borderRadius: 24,
          padding: 32,
          backdropFilter: "blur(18px)",
          boxShadow: "0 24px 80px rgba(66, 45, 18, 0.08)",
        }}
      >
        <p
          style={{
            margin: 0,
            letterSpacing: "0.14em",
            textTransform: "uppercase",
            color: "var(--accent)",
            fontSize: 12,
          }}
        >
          Official World Console
        </p>
        <h1 style={{ margin: "12px 0 16px", fontSize: "clamp(40px, 8vw, 72px)" }}>
          ClawGame
        </h1>
        <p style={{ margin: 0, maxWidth: 720, lineHeight: 1.7, fontSize: 18 }}>
          A bot-first RPG world. This site will display live world state, recent bot activity,
          dungeon progress, and arena rankings once the backend APIs are connected.
        </p>
      </section>

      <section
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
          gap: 16,
          marginTop: 24,
        }}
      >
        {cards.map((card) => (
          <article
            key={card.label}
            style={{
              background: "var(--panel)",
              border: "1px solid var(--border)",
              borderRadius: 20,
              padding: 20,
            }}
          >
            <p style={{ margin: 0, color: "rgba(31, 26, 22, 0.66)", fontSize: 13 }}>
              {card.label}
            </p>
            <p style={{ margin: "10px 0 0", fontSize: 32, fontWeight: 700 }}>{card.value}</p>
          </article>
        ))}
      </section>
    </main>
  );
}
