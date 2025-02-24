-- DDL Queries: Database Schema Creation
DROP TABLE IF EXISTS admins CASCADE;

-- Table: Store Admins
CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    jwt_token TEXT, -- Added JWT Token
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample store admins (sample seed)
INSERT INTO admins (name, email, password) VALUES
('Store Admin Alice', 'admin.alice@ecostore.com', 'adminpass123'),
('Store Admin Bob', 'admin.bob@greenshop.com', 'adminpass456');