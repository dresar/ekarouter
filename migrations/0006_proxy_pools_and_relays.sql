ALTER TABLE proxy_profiles ADD COLUMN no_proxy TEXT DEFAULT '';
ALTER TABLE proxy_profiles ADD COLUMN strict_proxy INTEGER NOT NULL DEFAULT 0;
ALTER TABLE proxy_profiles ADD COLUMN relay_type TEXT NOT NULL DEFAULT 'standard';
ALTER TABLE proxy_profiles ADD COLUMN relay_config TEXT DEFAULT '';
