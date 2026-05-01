CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO products (id, name, description, price, stock) VALUES
    ('prod-001', 'Go Programming Book', 'Learn Go programming', 39.99, 100),
    ('prod-002', 'Docker Deep Dive', 'Master Docker containers', 49.99, 50),
    ('prod-003', 'Kubernetes in Action', 'K8s for production', 59.99, 75);
