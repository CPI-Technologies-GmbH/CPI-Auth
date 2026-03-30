-- Add locale column to users table for email localization
ALTER TABLE users ADD COLUMN IF NOT EXISTS locale VARCHAR(10) NOT NULL DEFAULT 'en';
