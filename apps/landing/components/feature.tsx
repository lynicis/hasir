interface FeatureItem {
  title: string;
  desc: string;
  spec: string;
}

const FEATURES: FeatureItem[] = [
  {
    title: "Git-over-SSH push",
    desc: "Push .proto files with plain git push over a custom SSH server built for schemas.",
    spec: "ssh :2222",
  },
  {
    title: "Automatic SDK generation",
    desc: "Every push validates and packages client SDKs for TypeScript, JavaScript, and Go.",
    spec: "ts, js, go",
  },
  {
    title: "Buf-native validation",
    desc: "Definitions are checked for breaking changes and lint rules before they land.",
    spec: "buf breaking",
  },
  {
    title: "Postgres-backed registry",
    desc: "Every schema version is stored, indexed, and queryable from a single source of truth.",
    spec: "postgresql",
  },
  {
    title: "Org & repo management",
    desc: "Manage organizations, repositories, and permissions from one dashboard.",
    spec: "role-based",
  },
  {
    title: "SSH key authentication",
    desc: "Developers authenticate with their own SSH keys, with no shared tokens and no leaks.",
    spec: "ed25519 / rsa",
  },
];

export function FeaturesSection() {
  return (
    <section id="features" className="hl-section--ruled">
      <div className="container">
        <div className="hl-rows-layout">
          <div className="hl-sticky-head">
            <h2>Built around Git-SSH, buf, and Postgres.</h2>
            <p className="lead">
              For teams who treat protobuf definitions as source of truth.
            </p>
          </div>

          <ul className="hl-spec-list">
            {FEATURES.map((feature) => (
              <li key={feature.title} className="hl-spec-row">
                <div className="hl-spec-row__body">
                  <h3>{feature.title}</h3>
                  <p>{feature.desc}</p>
                </div>
                <span className="hl-spec-row__token">{feature.spec}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
