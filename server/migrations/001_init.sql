-- Initial schema for orderfood (reference; GORM AutoMigrate also applies)

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    open_id VARCHAR(64) NOT NULL UNIQUE,
    nickname VARCHAR(64) DEFAULT '',
    avatar_url VARCHAR(512) DEFAULT '',
    role VARCHAR(16) NOT NULL DEFAULT 'buyer',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS buyer_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    contact_name VARCHAR(64) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    address VARCHAR(512) NOT NULL,
    address_lat DOUBLE NULL,
    address_lng DOUBLE NULL,
    profile_completed TINYINT(1) NOT NULL DEFAULT 0,
    updated_at DATETIME(3) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS seller_phones (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone VARCHAR(20) NOT NULL UNIQUE,
    shop_name VARCHAR(128) NOT NULL DEFAULT '我的店铺',
    created_at DATETIME(3) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS seller_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    shop_name VARCHAR(128) NOT NULL,
    address VARCHAR(512) NOT NULL,
    address_lat DOUBLE NOT NULL,
    address_lng DOUBLE NOT NULL,
    updated_at DATETIME(3) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS menus (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    seller_id BIGINT UNSIGNED NOT NULL,
    delivery_date DATE NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_menus_seller_id (seller_id),
    INDEX idx_menus_delivery_date (delivery_date),
    UNIQUE INDEX uk_menus_seller_delivery (seller_id, delivery_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS menu_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    menu_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(128) NOT NULL,
    image_url VARCHAR(512) DEFAULT '',
    price BIGINT NOT NULL,
    description VARCHAR(512) DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    is_available TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_menu_items_menu_id (menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL UNIQUE,
    buyer_id BIGINT UNSIGNED NOT NULL,
    menu_id BIGINT UNSIGNED NOT NULL,
    delivery_date DATE NOT NULL,
    total_amount BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_payment',
    contact_name VARCHAR(64) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    address VARCHAR(512) NOT NULL,
    address_lat DOUBLE NULL,
    address_lng DOUBLE NULL,
    delivery_time_pref VARCHAR(64) DEFAULT '',
    remark VARCHAR(512) DEFAULT '',
    refund_reason VARCHAR(256) DEFAULT '',
    refund_remark VARCHAR(512) DEFAULT '',
    refunded_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_orders_buyer_id (buyer_id),
    INDEX idx_orders_delivery_date (delivery_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    menu_item_id BIGINT UNSIGNED NOT NULL,
    name_snapshot VARCHAR(128) NOT NULL,
    price_snapshot BIGINT NOT NULL,
    quantity INT NOT NULL,
    created_at DATETIME(3) NOT NULL,
    INDEX idx_order_items_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL UNIQUE,
    transaction_id VARCHAR(64) DEFAULT '',
    amount BIGINT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    paid_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS cutoff_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_date DATE NOT NULL UNIQUE,
    cutoff_time VARCHAR(8) NOT NULL,
    updated_at DATETIME(3) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS order_messages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    sender_id BIGINT UNSIGNED NOT NULL,
    sender_role VARCHAR(16) NOT NULL,
    type VARCHAR(16) NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME(3) NOT NULL,
    INDEX idx_order_messages_order_id (order_id),
    INDEX idx_order_messages_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS delivery_routes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    seller_id BIGINT UNSIGNED NOT NULL,
    delivery_date DATE NOT NULL,
    stops_json TEXT NOT NULL,
    total_distance INT NOT NULL DEFAULT 0,
    total_duration INT NOT NULL DEFAULT 0,
    updated_at DATETIME(3) NOT NULL,
    UNIQUE INDEX uk_delivery_routes_seller_date (seller_id, delivery_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
