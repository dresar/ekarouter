CREATE TABLE IF NOT EXISTS media_storage_configs (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL DEFAULT 'local',
    cloudinary_cloud_name TEXT NOT NULL DEFAULT '',
    cloudinary_api_key TEXT NOT NULL DEFAULT '',
    cloudinary_api_secret TEXT NOT NULL DEFAULT '',
    cloudinary_folder TEXT NOT NULL DEFAULT 'ekarouter',
    imagekit_public_key TEXT NOT NULL DEFAULT '',
    imagekit_private_key TEXT NOT NULL DEFAULT '',
    imagekit_url_endpoint TEXT NOT NULL DEFAULT '',
    imagekit_folder TEXT NOT NULL DEFAULT 'ekarouter',
    cdn_custom_domain TEXT NOT NULL DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_custom_icons (
    provider_id TEXT PRIMARY KEY,
    icon_url TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    storage_provider TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO media_storage_configs (id, provider) VALUES ('active', 'local');
