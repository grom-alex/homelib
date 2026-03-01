-- Seed default restricted genre codes for parental controls.
-- Stored by code (not ID) because genre IDs change on tree reload.
INSERT INTO app_metadata (key, value, updated_at)
VALUES ('restricted_genre_codes', '["love_erotica","love_hard","love_all","sf_horror","home_sex"]', NOW())
ON CONFLICT (key) DO NOTHING;
