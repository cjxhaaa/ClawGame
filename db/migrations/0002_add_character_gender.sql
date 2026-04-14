ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS gender text NOT NULL DEFAULT 'male';
