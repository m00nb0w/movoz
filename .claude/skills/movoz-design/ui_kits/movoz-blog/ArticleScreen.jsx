// Movoz blog — single article reading view
const { Badge, Avatar, PlaceholderBox, Button, Card, Annotation, Skeleton } = window.MovozDesignSystem_c2bc1d;

function ArticleScreen({ annotate, onBack }) {
  return (
    <article style={{ background: 'var(--paper)' }}>
      <div style={{ maxWidth: 760, margin: '0 auto', padding: '40px 24px 80px' }}>
        <button onClick={onBack} style={{
          fontFamily: 'var(--font-marker)', fontStyle: 'italic', fontSize: 16,
          color: 'var(--ink-soft)', background: 'none', border: 'none', cursor: 'pointer', padding: 0, marginBottom: 22,
        }}>← All articles</button>

        <div style={{ display: 'flex', gap: 8, marginBottom: 18 }}>
          <Badge tone="accent">Featured</Badge>
          <Badge>Token Security</Badge>
        </div>

        <h1 style={{ fontSize: 46, lineHeight: 1.06, margin: '0 0 20px' }}>
          How we built token rotation that survives 50M sessions
        </h1>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 28 }}>
          <Avatar initials="MV" size={44} />
          <div>
            <div style={{ fontFamily: 'var(--font-marker)', fontWeight: 600, fontSize: 17 }}>M. Vance</div>
            <div className="movoz-label">Jun 22, 2026 · 14 min read</div>
          </div>
        </div>

        {annotate && <Annotation style={{ marginBottom: 10 }}>lead figure — token exchange</Annotation>}
        <PlaceholderBox label="token exchange diagram" ratio="16 / 9" style={{ marginBottom: 32 }} />

        <div style={{ fontFamily: 'var(--font-sans)', fontSize: 18, lineHeight: 1.6, color: 'var(--ink-2)' }}>
          <p>Movoz issues short-lived access tokens and rotates refresh tokens on every exchange. When a previously-used refresh token reappears, the whole token family is revoked — reuse becomes a signal, not a vulnerability.</p>
          <Skeleton lines={4} style={{ margin: '20px 0' }} />
          <h2 style={{ fontSize: 28, margin: '36px 0 14px' }}>Rotation under load</h2>
          <Skeleton lines={5} style={{ margin: '20px 0' }} />
        </div>

        <Card lift style={{ margin: '34px 0', background: 'var(--terracotta-faint)', borderColor: 'var(--terracotta)' }}>
          <div className="movoz-label" style={{ color: 'var(--terracotta-ink)' }}>Key takeaway</div>
          <p style={{ fontFamily: 'var(--font-marker)', fontSize: 22, margin: '8px 0 0', color: 'var(--ink)' }}>
            Rotate on every exchange. Detect reuse. Revoke the family.
          </p>
        </Card>

        <div style={{ fontFamily: 'var(--font-sans)', fontSize: 18, lineHeight: 1.6, color: 'var(--ink-2)' }}>
          <Skeleton lines={4} />
        </div>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 44, paddingTop: 24, borderTop: 'var(--border-width) solid var(--line)' }}>
          <span className="movoz-label">Next · OAuth 2.1</span>
          <Button variant="secondary" iconRight={<span>→</span>}>Anatomy of an access token</Button>
        </div>
      </div>
    </article>
  );
}
window.ArticleScreen = ArticleScreen;
