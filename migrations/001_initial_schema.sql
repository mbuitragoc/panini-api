CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  apple_sub TEXT UNIQUE NOT NULL,
  username TEXT UNIQUE,
  handle TEXT UNIQUE,
  device_token TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE stickers (
  id TEXT PRIMARY KEY,
  country_code TEXT NOT NULL,
  sticker_number INTEGER NOT NULL,
  type TEXT NOT NULL,
  player_name TEXT,
  dob DATE,
  height REAL,
  weight REAL,
  club TEXT,
  club_country TEXT,
  position TEXT,
  national_team TEXT NOT NULL,
  image_url TEXT
);

CREATE TABLE user_collections (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  sticker_id TEXT NOT NULL REFERENCES stickers(id),
  quantity_owned INTEGER NOT NULL DEFAULT 0,
  wishlisted BOOLEAN NOT NULL DEFAULT FALSE,
  blacklisted BOOLEAN NOT NULL DEFAULT FALSE,
  first_acquired_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, sticker_id)
);

CREATE TABLE friendships (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, friend_id),
  CHECK (user_id <> friend_id)
);

CREATE TABLE trades (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  proposer_id UUID NOT NULL REFERENCES users(id),
  recipient_id UUID NOT NULL REFERENCES users(id),
  status TEXT NOT NULL DEFAULT 'proposed',
  offered_stickers TEXT[] NOT NULL DEFAULT '{}',
  requested_stickers TEXT[] NOT NULL DEFAULT '{}',
  proposed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_collections_user_id ON user_collections(user_id);
CREATE INDEX idx_user_collections_updated_at ON user_collections(updated_at);
CREATE INDEX idx_friendships_friend_id ON friendships(friend_id);
CREATE INDEX idx_trades_proposer_id ON trades(proposer_id);
CREATE INDEX idx_trades_recipient_id ON trades(recipient_id);
CREATE INDEX idx_trades_updated_at ON trades(updated_at);
