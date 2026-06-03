-- =============================================================
-- fx 项目 - MySQL 建表 DDL
-- 使用方式: mysql -h <host> -u <user> -p < dbname < schema/mysql.sql
-- 也可直接用 GORM AutoMigrate（项目启动时自动执行）
-- =============================================================

CREATE TABLE IF NOT EXISTS `users` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username`   VARCHAR(50)     NOT NULL,
    `email`      VARCHAR(100)    NOT NULL,
    `password`   VARCHAR(255)    NOT NULL,
    `created_at` DATETIME(3)     NOT NULL,
    `updated_at` DATETIME(3)     NOT NULL,
    `deleted_at` DATETIME(3)     NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_username` (`username`),
    UNIQUE INDEX `idx_users_email` (`email`),
    INDEX `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `refresh_tokens` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT UNSIGNED NOT NULL,
    `token`      VARCHAR(500)    NOT NULL,
    `expires_at` DATETIME(3)     NOT NULL,
    `created_at` DATETIME(3)     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_refresh_tokens_token` (`token`),
    INDEX `idx_refresh_tokens_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `trades` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    `uuid`       VARCHAR(36)      NOT NULL,
    `user_id`    BIGINT UNSIGNED  NOT NULL,
    `asset_type` VARCHAR(20)      NOT NULL DEFAULT 'crypto',
    `symbol`     VARCHAR(20)      NOT NULL,
    `type`       VARCHAR(10)      NOT NULL COMMENT 'buy / sell / recharge',
    `amount`     DECIMAL(20,8)    NOT NULL,
    `price`      DECIMAL(20,8)    NOT NULL,
    `total`      DECIMAL(20,8)    NOT NULL,
    `currency`   VARCHAR(10)      NOT NULL DEFAULT 'USD',
    `created_at` DATETIME(3)      NOT NULL,
    `updated_at` DATETIME(3)      NOT NULL,
    `deleted_at` DATETIME(3)      NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_trades_uuid` (`uuid`),
    INDEX `idx_trades_user_id` (`user_id`),
    INDEX `idx_trades_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `holdings` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT UNSIGNED  NOT NULL,
    `asset_type` VARCHAR(20)      NOT NULL DEFAULT 'crypto',
    `symbol`     VARCHAR(20)      NOT NULL,
    `amount`     DECIMAL(20,8)    NOT NULL,
    `currency`   VARCHAR(10)      NOT NULL DEFAULT 'USD',
    `created_at` DATETIME(3)      NOT NULL,
    `updated_at` DATETIME(3)      NOT NULL,
    `deleted_at` DATETIME(3)      NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_holdings_user_id` (`user_id`),
    INDEX `idx_holdings_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `exchange_rates` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    `from`       VARCHAR(10)      NOT NULL COMMENT '源币种',
    `to`         VARCHAR(10)      NOT NULL COMMENT '目标币种',
    `rate`       DECIMAL(20,8)    NOT NULL,
    `updated_at` DATETIME(3)      NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_exchange_rates_from_to` (`from`, `to`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
