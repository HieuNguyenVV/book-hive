-- Drop timestamp defaults before changing column types
ALTER TABLE "users" ALTER COLUMN "deleted_at" DROP DEFAULT;
ALTER TABLE "users" ALTER COLUMN "created_at" DROP DEFAULT;
ALTER TABLE "users" ALTER COLUMN "updated_at" DROP DEFAULT;

ALTER TABLE "tokens" ALTER COLUMN "created_at" DROP DEFAULT;
ALTER TABLE "tokens" ALTER COLUMN "updated_at" DROP DEFAULT;

-- Change timestamp columns to unix epoch (BIGINT)
ALTER TABLE "users"
  ALTER COLUMN "last_login_at" TYPE bigint USING EXTRACT(EPOCH FROM "last_login_at")::bigint,
  ALTER COLUMN "deleted_at" TYPE bigint USING EXTRACT(EPOCH FROM "deleted_at")::bigint,
  ALTER COLUMN "created_at" TYPE bigint USING EXTRACT(EPOCH FROM "created_at")::bigint,
  ALTER COLUMN "updated_at" TYPE bigint USING EXTRACT(EPOCH FROM "updated_at")::bigint;

ALTER TABLE "tokens"
  ALTER COLUMN "revoked_at" TYPE bigint USING EXTRACT(EPOCH FROM "revoked_at")::bigint,
  ALTER COLUMN "expires_at" TYPE bigint USING EXTRACT(EPOCH FROM "expires_at")::bigint,
  ALTER COLUMN "created_at" TYPE bigint USING EXTRACT(EPOCH FROM "created_at")::bigint,
  ALTER COLUMN "updated_at" TYPE bigint USING EXTRACT(EPOCH FROM "updated_at")::bigint;

-- Set bigint defaults
ALTER TABLE "users" ALTER COLUMN "deleted_at" SET DEFAULT 0;
ALTER TABLE "users" ALTER COLUMN "created_at" SET DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint;
ALTER TABLE "users" ALTER COLUMN "updated_at" SET DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint;

ALTER TABLE "tokens" ALTER COLUMN "created_at" SET DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint;
ALTER TABLE "tokens" ALTER COLUMN "updated_at" SET DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint;
