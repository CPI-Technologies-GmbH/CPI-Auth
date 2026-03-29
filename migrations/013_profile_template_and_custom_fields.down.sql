DELETE FROM page_templates WHERE id = 'f0000001-0000-0000-0000-000000000008';
DELETE FROM template_language_strings WHERE string_key LIKE 'profile.%';
DELETE FROM template_language_strings WHERE string_key = 'signup.custom_fields_heading';
