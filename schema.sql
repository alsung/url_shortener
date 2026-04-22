-- schema.sql

CREATE TABLE links (
    id          BIGSERIAL PRIMARY KEY,
    short_code  VARCHAR(10)  NOT NULL UNIQUE,
    long_url    TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_links_short_code ON links(short_code);

CREATE TABLE clicks (
    id          BIGSERIAL PRIMARY KEY,
    short_code  VARCHAR(10)  NOT NULL,
    clicked_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    referrer    TEXT
);

-- Partial index: only index non-null referrers (saves space, still fast for stats queries)
CREATE INDEX idx_clicks_short_code ON clicks(short_code);
CREATE INDEX idx_clicks_referrer ON clicks(referrer) WHERE referrer IS NOT NULL;