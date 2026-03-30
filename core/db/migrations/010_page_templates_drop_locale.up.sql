-- Make page templates locale-independent
ALTER TABLE page_templates DROP COLUMN locale;
ALTER TABLE page_templates DROP CONSTRAINT IF EXISTS page_templates_tenant_id_page_type_locale_key;
ALTER TABLE page_templates ADD CONSTRAINT page_templates_tenant_id_page_type_key UNIQUE (tenant_id, page_type);
