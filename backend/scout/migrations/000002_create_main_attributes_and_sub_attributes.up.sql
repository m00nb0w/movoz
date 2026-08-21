CREATE TABLE IF NOT EXISTS main_attributes (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sub_attributes (
    id SERIAL PRIMARY KEY,
    main_attribute_id INTEGER NOT NULL REFERENCES main_attributes(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sub_attributes_main_attribute_id ON sub_attributes(main_attribute_id);
CREATE INDEX idx_sub_attributes_is_active ON sub_attributes(is_active);

INSERT INTO main_attributes (key, name) VALUES
    ('technical_expertise', 'Technical Expertise'),
    ('critical_thinking', 'Critical Thinking'),
    ('communication', 'Communication'),
    ('management', 'Management'),
    ('product_mindset', 'Product Mindset'),
    ('force_multiplier', 'Force Multiplier');
