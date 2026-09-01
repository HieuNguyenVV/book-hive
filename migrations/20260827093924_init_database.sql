-- Create enum type "user_role"
CREATE TYPE "user_role" AS ENUM ('admin', 'user', 'guest');
-- Create enum type "user_status"
CREATE TYPE "user_status" AS ENUM ('active', 'inactive', 'pending', 'deleted');
-- Create "users" table
CREATE TABLE "users" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "email" character varying(255) NOT NULL, "password" text NOT NULL, "first_name" character varying(255) NOT NULL, "last_name" character varying(255) NOT NULL, "full_name" character varying(255) NOT NULL, "role" "user_role" NOT NULL, "status" "user_status" NOT NULL, "last_login_at" timestamp NULL, "deleted_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("id"), CONSTRAINT "users_email_key" UNIQUE ("email"));
-- Create index "idx_users_email" to table: "users"
CREATE INDEX "idx_users_email" ON "users" ("email");
-- Create index "idx_users_full_name" to table: "users"
CREATE INDEX "idx_users_full_name" ON "users" ("full_name");
-- Create index "idx_users_name" to table: "users"
CREATE INDEX "idx_users_name" ON "users" ("first_name", "last_name");
-- Create index "idx_users_role" to table: "users"
CREATE INDEX "idx_users_role" ON "users" ("role");
-- Create index "idx_users_status" to table: "users"
CREATE INDEX "idx_users_status" ON "users" ("status");
-- Create "tokens" table
CREATE TABLE "tokens" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "user_id" uuid NOT NULL, "token" text NOT NULL, "revoked_at" timestamp NULL, "expires_at" timestamp NOT NULL, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("id"), CONSTRAINT "tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "idx_tokens_token" to table: "tokens"
CREATE INDEX "idx_tokens_token" ON "tokens" ("token");
-- Create index "idx_tokens_user_id" to table: "tokens"
CREATE INDEX "idx_tokens_user_id" ON "tokens" ("user_id");
