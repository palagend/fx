-- =============================================================
-- fx 项目 - PostgreSQL 建表 DDL
-- 使用方式: psql -h <host> -U <user> -d <dbname> -f schema/postgres.sql
-- 也可直接用 GORM AutoMigrate（项目启动时自动执行）
-- =============================================================

CREATE TABLE IF NOT EXISTS "users" (
    "id"         BIGSERIAL    NOT NULL,
    "username"   VARCHAR(50)  NOT NULL,
    "email"      VARCHAR(100) NOT NULL,
    "password"   VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ  NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "idx_users_username" UNIQUE ("username"),
    CONSTRAINT "idx_users_email"    UNIQUE ("email")
);
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");

CREATE TABLE IF NOT EXISTS "refresh_tokens" (
    "id"         BIGSERIAL    NOT NULL,
    "user_id"    BIGINT       NOT NULL,
    "token"      VARCHAR(500) NOT NULL,
    "expires_at" TIMESTAMPTZ  NOT NULL,
    "created_at" TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("id"),
    CONSTRAINT "idx_refresh_tokens_token" UNIQUE ("token")
);
CREATE INDEX IF NOT EXISTS "idx_refresh_tokens_user_id" ON "refresh_tokens" ("user_id");

CREATE TABLE IF NOT EXISTS "trades" (
    "id"         BIGSERIAL      NOT NULL,
    "uuid"       VARCHAR(36)    NOT NULL,
    "user_id"    BIGINT         NOT NULL,
    "asset_type" VARCHAR(20)    NOT NULL DEFAULT 'crypto',
    "symbol"     VARCHAR(20)    NOT NULL,
    "type"       VARCHAR(10)    NOT NULL,
    "amount"     NUMERIC(20,8)  NOT NULL,
    "price"      NUMERIC(20,8)  NOT NULL,
    "total"      NUMERIC(20,8)  NOT NULL,
    "currency"   VARCHAR(10)    NOT NULL DEFAULT 'USD',
    "created_at" TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ    NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "idx_trades_uuid" UNIQUE ("uuid")
);
CREATE INDEX IF NOT EXISTS "idx_trades_user_id"    ON "trades" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_trades_deleted_at" ON "trades" ("deleted_at");

CREATE TABLE IF NOT EXISTS "holdings" (
    "id"         BIGSERIAL      NOT NULL,
    "user_id"    BIGINT         NOT NULL,
    "asset_type" VARCHAR(20)    NOT NULL DEFAULT 'crypto',
    "symbol"     VARCHAR(20)    NOT NULL,
    "amount"     NUMERIC(20,8)  NOT NULL,
    "currency"   VARCHAR(10)    NOT NULL DEFAULT 'USD',
    "created_at" TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ    NULL,
    PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_holdings_user_id"    ON "holdings" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_holdings_deleted_at" ON "holdings" ("deleted_at");

CREATE TABLE IF NOT EXISTS "exchange_rates" (
    "id"         BIGSERIAL      NOT NULL,
    "from"       VARCHAR(10)    NOT NULL,
    "to"         VARCHAR(10)    NOT NULL,
    "rate"       NUMERIC(20,8)  NOT NULL,
    "updated_at" TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_exchange_rates_from_to" ON "exchange_rates" ("from", "to");
