DELETE FROM page_templates WHERE is_default = TRUE;
ALTER TABLE page_templates DROP COLUMN IF EXISTS is_default;
DROP TABLE IF EXISTS template_language_strings;
