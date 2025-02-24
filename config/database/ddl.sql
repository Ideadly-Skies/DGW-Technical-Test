-- DDL Queries: Database Schema Creation
DROP TABLE IF EXISTS admins CASCADE;
DROP TABLE IF EXISTS products CASCADE;

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

-- Table: Products
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),         
    admin_id UUID REFERENCES admins(id) ON DELETE SET NULL, 
    name VARCHAR(255) NOT NULL,                             
    description TEXT,                                      
    price DECIMAL(10, 2) NOT NULL,                          
    stock_quantity INT NOT NULL,                            
    category VARCHAR(255),                                  
    image_url TEXT,                                         
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,         
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP          
);

-- Insert sample store admins (sample seed)
INSERT INTO admins (name, email, password) VALUES
('Store Admin Alice', 'admin.alice@ecostore.com', 'adminpass123'),
('Store Admin Bob', 'admin.bob@greenshop.com', 'adminpass456');

-- Insert sample products (sample seed)
INSERT INTO products (admin_id, name, description, price, stock_quantity, category, image_url) VALUES
((SELECT id FROM admins WHERE email = 'admin.alice@ecostore.com'), 'Eco-friendly T-Shirt', 'A comfortable eco-friendly t-shirt made from organic cotton.', 19.99, 100, 'Apparel', 'https://ecostore.com/images/t-shirt.jpg'),
((SELECT id FROM admins WHERE email = 'admin.bob@greenshop.com'), 'Green Energy Light Bulb', 'LED light bulb with energy-saving features.', 5.99, 200, 'Electronics', 'https://greenshop.com/images/lightbulb.jpg');