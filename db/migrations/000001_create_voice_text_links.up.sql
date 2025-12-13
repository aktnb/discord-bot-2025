CREATE TABLE voice_text_links (
    id TEXT PRIMARY KEY,
    voice_channel_id TEXT NOT NULL,
    guild_id TEXT NOT NULL,
    text_channel_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (guild_id, voice_channel_id)
);

CREATE INDEX idx_voice_text_links_guild_voice
    ON voice_text_links (guild_id, voice_channel_id);