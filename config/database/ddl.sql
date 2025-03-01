-- DDL Queries: Schema Creation (9)
-- Drop the dependent tables first (those that reference other tables)
DROP TABLE IF EXISTS wallet_transactions CASCADE;  
DROP TABLE IF EXISTS reviews CASCADE;              
DROP TABLE IF EXISTS logs CASCADE;                 
DROP TABLE IF EXISTS payments CASCADE;             
DROP TABLE IF EXISTS order_items CASCADE;          
DROP TABLE IF EXISTS orders CASCADE;               
DROP TABLE IF EXISTS products CASCADE;             
DROP TABLE IF EXISTS suppliers CASCADE;            
DROP TABLE IF EXISTS admins CASCADE;              
DROP TABLE IF EXISTS farmers CASCADE;

-- Table: Admins (DGW)
CREATE TABLE admins (
    id SERIAL PRIMARY KEY, 
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    role VARCHAR(100) NOT NULL,
    jwt_token TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: Farmers
CREATE TABLE farmers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    address VARCHAR(500),
    phone_number VARCHAR(100),
    farm_type VARCHAR(100),
    wallet_balance DECIMAL(10, 2) DEFAULT 0.00,
    jwt_token TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: Wallet Transactions
CREATE TABLE wallet_transactions (
    id SERIAL PRIMARY KEY,
    farmer_id INTEGER REFERENCES farmers(id) ON DELETE CASCADE,
    order_id VARCHAR(255) NOT NULL,  -- Add order_id field
    transaction_type VARCHAR(100),
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) CHECK (status IN ('pending', 'settlement', 'failed')) DEFAULT 'pending',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Add created_at column
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- Add updated_at column
);

-- Table: Suppliers
CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address VARCHAR(100),
    phone_number VARCHAR(100),
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: Products
CREATE TABLE products (
    id SERIAL PRIMARY KEY,         
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE CASCADE, 
    name VARCHAR(250) NOT NULL,                             
    description TEXT,                                      
    price DECIMAL(10, 2) NOT NULL,                          
    stock_quantity INT NOT NULL,                            
    category VARCHAR(100),                                  
    brand VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,         
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP          
);

-- Table: Orders
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    farmer_id INTEGER REFERENCES farmers(id) ON DELETE CASCADE,
    status VARCHAR(100) CHECK (status IN ('pending', 'settlement', 'cancelled')),
    total_price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: Order Items
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    quantity INT,
    price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: Payments
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2),
    payment_method VARCHAR(250),
    status VARCHAR(100) CHECK (status IN ('pending', 'settlement', 'failed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: Logs
CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    admin_id INTEGER REFERENCES admins(id) ON DELETE CASCADE,
    action VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    details TEXT
);

-- Table: Reviews
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    farmer_id INTEGER REFERENCES farmers(id) ON DELETE CASCADE,
    rating INT CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(100) CHECK (status IN ('pending', 'approved', 'rejected'))
);

-- DML Query: populating the created schemas
-- Insert sample admins
INSERT INTO admins (name, email, password, role) VALUES
('Super Admin Diana', 'superadmin@dgwmart.com', 'adminpass123', 'Super Admin'),
('Admin Adi', 'admin.adi@dgwmart.com', 'adminpass456', 'Store Admin');

-- Insert sample suppliers
INSERT INTO suppliers (name, address, phone_number, category) VALUES
('PT Dharma Guna Wibawa', 'Jl. Raya Bogor No. 123, Jakarta', '08123456789', 'Agrokimia'),
('DGW Fertilizer', 'Jl. Gresik No. 5, Gresik, Jawa Timur', '08567890123', 'Pupuk'),
('PT Semesta Alam Sejati', 'Jl. Alam Sejati No. 10, Surabaya', '08987654321', 'Alat Pertanian'),
('PT Bangun Sahabat Tani', 'Jl. Distribusi No. 15, Jakarta', '08123456788', 'Distribusi Internal');

-- Insert sample products
INSERT INTO products (supplier_id, name, description, price, stock_quantity, category, brand) VALUES
(1, 'Supremo Herbicide', 'Herbicide for weed control', 50000, 100, 'Agrokimia', 'DGW'),
(1, 'Klensect Pesticide', 'Effective pesticide for pest control', 65000, 120, 'Agrokimia', 'DGW'),
(2, 'Premium NPK Fertilizer', 'High-quality NPK fertilizer', 75000, 150, 'Pupuk', 'DGW'),
(2, 'Organic NPK Fertilizer', 'Organic-based NPK fertilizer', 1500000, 200, 'Pupuk', 'DGW'),
(3, 'Electric Sprayer', 'Battery-operated sprayer for efficient spraying', 20000, 100, 'Alat Pertanian', 'SAS'),
(3, 'Manual Sprayer', 'Manual sprayer for small-scale farming', 20000, 300, 'Alat Pertanian', 'SAS'),
(3, 'Plastic Mulsa', 'Premium quality plastic mulch for crop protection', 15000, 500, 'Alat Pertanian', 'SAS');