// Movoz blog — top navigation bar
const { Logo, Input, Button } = window.MovozDesignSystem_c2bc1d;

function SearchIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
      <circle cx="11" cy="11" r="7" /><line x1="21" y1="21" x2="16.5" y2="16.5" />
    </svg>
  );
}

function NavBar({ onNav, active = 'Articles' }) {
  const links = ['Articles', 'Topics', 'Open Source', 'Careers'];
  return (
    <header style={{
      display: 'flex', alignItems: 'center', gap: 24,
      padding: '18px 40px',
      borderBottom: 'var(--border-width) solid var(--line)',
      background: 'var(--paper)',
      position: 'sticky', top: 0, zIndex: 20,
    }}>
      <div style={{ cursor: 'pointer' }} onClick={() => onNav && onNav('Home')}>
        <Logo sublabel="logo + wordmark" />
      </div>
      <nav style={{ display: 'flex', gap: 26, marginLeft: 28 }}>
        {links.map(l => (
          <a key={l} onClick={() => onNav && onNav(l)} style={{
            fontFamily: 'var(--font-marker)', fontStyle: 'italic',
            fontSize: 'var(--text-lg)', fontWeight: active === l ? 700 : 500,
            color: active === l ? 'var(--ink)' : 'var(--ink-soft)',
            textDecoration: 'none', cursor: 'pointer',
          }}>{l}</a>
        ))}
      </nav>
      <div style={{ marginLeft: 'auto', display: 'flex', gap: 12, alignItems: 'center' }}>
        <div style={{ width: 260 }}>
          <Input pill placeholder="Search…" iconLeft={<SearchIcon />} />
        </div>
        <Button variant="secondary">Subscribe</Button>
      </div>
    </header>
  );
}
window.NavBar = NavBar;
window.SearchIcon = SearchIcon;
