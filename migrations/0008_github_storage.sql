ALTER TABLE media_storage_configs ADD COLUMN github_token TEXT NOT NULL DEFAULT '';
ALTER TABLE media_storage_configs ADD COLUMN github_owner TEXT NOT NULL DEFAULT '';
ALTER TABLE media_storage_configs ADD COLUMN github_repo TEXT NOT NULL DEFAULT '';
ALTER TABLE media_storage_configs ADD COLUMN github_branch TEXT NOT NULL DEFAULT 'main';
ALTER TABLE media_storage_configs ADD COLUMN github_folder TEXT NOT NULL DEFAULT 'uploads';
ALTER TABLE media_storage_configs ADD COLUMN github_cdn_domain TEXT NOT NULL DEFAULT '';
