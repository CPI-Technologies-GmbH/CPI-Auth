ALTER TABLE page_templates DROP CONSTRAINT IF EXISTS page_templates_tenant_id_page_type_key;
ALTER TABLE page_templates ADD COLUMN locale VARCHAR(10) NOT NULL DEFAULT 'en';
ALTER TABLE page_templates ADD CONSTRAINT page_templates_tenant_id_page_type_locale_key UNIQUE (tenant_id, page_type, locale);
