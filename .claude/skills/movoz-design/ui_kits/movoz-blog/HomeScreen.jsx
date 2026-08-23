// Movoz blog — home / index view (hero + filter + article grid)
const { Pill, Card, Button, Badge, Avatar, PlaceholderBox, Annotation, Skeleton } = window.MovozDesignSystem_c2bc1d;

const MOVOZ_POSTS = [
  { cat: 'Architecture', min: 9, title: 'OIDC discovery at the edge', author: 'R. Okafor' },
  { cat: 'Token Security', min: 12, title: 'Refresh-token reuse detection', author: 'L. Park' },
  { cat: 'OAuth 2.1', min: 7, title: 'PKCE by default', author: 'M. Vance' },
  { cat: 'Token Security', min: 11, title: 'Sender-constrained tokens with DPoP', author: 'S. Adeyemi' },
  { cat: 'Ways of Working', min: 6, title: 'A zero-meeting RFC process', author: 'J. Wu' },
  { cat: 'OAuth 2.1', min: 8, title: 'Anatomy of an access token', author: 'R. Okafor' },
];

function MetaLabel({ children }) {
  return <span className="movoz-label">{children}</span>;
}

function HomeScreen({ filter, setFilter, annotate, onOpen }) {
  const filters = ['All', 'OAuth 2.1', 'OpenID Connect', 'Token Security', 'Architecture', 'Ways of Working'];
  const posts = filter === 'All' ? MOVOZ_POSTS : MOVOZ_POSTS.filter(p => p.cat === filter);

  return (
    <div>
      {/* Hero band */}
      <section style={{ background: 'var(--paper-sunken)', borderBottom: 'var(--border-width) solid var(--line)' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', padding: '28px 40px 48px', position: 'relative' }}>
          {annotate && <Annotation style={{ marginBottom: 18 }}>hero — featured / latest post</Annotation>}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 56, alignItems: 'center', marginTop: annotate ? 0 : 14 }}>
            <div>
              <MetaLabel>Featured · Token Security · 14 min</MetaLabel>
              <h1 style={{ fontSize: 52, margin: '16px 0 18px', lineHeight: 1.05 }}>
                How we built token rotation that survives 50M sessions
              </h1>
              <div style={{ maxWidth: 440 }}><Skeleton lines={3} /></div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '24px 0 28px' }}>
                <Avatar initials="MV" size={40} />
                <div><Skeleton width={180} height={11} /></div>
              </div>
              <Button variant="primary" lift iconRight={<span>→</span>} onClick={() => onOpen && onOpen()}>Read article</Button>
            </div>
            <PlaceholderBox label="code sample / token exchange" ratio="5 / 4" />
          </div>
        </div>
      </section>

      {/* Filter row */}
      <section style={{ borderBottom: 'var(--border-width) solid var(--line)', background: 'var(--paper)' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', padding: '18px 40px', display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <span className="movoz-label" style={{ marginRight: 4 }}>Filter</span>
          {filters.map(f => <Pill key={f} active={filter === f} onClick={() => setFilter(f)}>{f}</Pill>)}
        </div>
      </section>

      {/* Article grid */}
      <section className="movoz-grid-bg" style={{ background: 'var(--paper)' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', padding: '36px 40px 80px' }}>
          <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 22 }}>
            <h2 style={{ fontSize: 34, margin: 0 }}>Latest from Movoz</h2>
            {annotate && <Annotation align="right">article grid · 3 cols</Annotation>}
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24 }}>
            {posts.map((p, i) => (
              <Card key={i} interactive padding="0" onClick={() => onOpen && onOpen()}>
                <div style={{ padding: 14 }}><PlaceholderBox label="cover image" ratio="16 / 9" /></div>
                <div style={{ padding: '4px 18px 20px' }}>
                  <MetaLabel>{p.cat} · {p.min} min</MetaLabel>
                  <h3 style={{ fontSize: 22, margin: '10px 0 12px' }}>{p.title}</h3>
                  <Skeleton lines={2} />
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 16 }}>
                    <Avatar size={28} />
                    <Skeleton width={120} height={10} />
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
window.HomeScreen = HomeScreen;
window.MetaLabel = MetaLabel;
