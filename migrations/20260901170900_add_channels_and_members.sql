-- Create "channels" table
CREATE TABLE "channels" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "channel_url" character varying(255) NOT NULL,
  "cover_url" text NULL,
  "name" character varying(255) NOT NULL,
  "channel_type" character varying(50) NOT NULL,
  "metadata" jsonb NOT NULL DEFAULT '{}',
  "last_message_id" uuid NULL,
  "last_read_message_id" uuid NULL,
  "is_public" boolean NOT NULL DEFAULT false,
  "create_by" uuid NOT NULL,
  "create_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  "update_by" uuid NOT NULL,
  "update_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  PRIMARY KEY ("id"),
  CONSTRAINT "channels_channel_url_key" UNIQUE ("channel_url"),
  CONSTRAINT "channels_create_by_fkey" FOREIGN KEY ("create_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "channels_update_by_fkey" FOREIGN KEY ("update_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX "idx_channels_channel_url" ON "channels" ("channel_url");
CREATE INDEX "idx_channels_channel_type" ON "channels" ("channel_type");
CREATE INDEX "idx_channels_is_public" ON "channels" ("is_public");
CREATE INDEX "idx_channels_create_by" ON "channels" ("create_by");
CREATE INDEX "idx_channels_name" ON "channels" ("name");

-- Create "members" table
CREATE TABLE "members" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "channel_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" character varying(50) NOT NULL,
  "joined_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  "inviter_id" uuid NULL,
  "created_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  "updated_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  PRIMARY KEY ("id"),
  CONSTRAINT "members_channel_user_key" UNIQUE ("channel_id", "user_id"),
  CONSTRAINT "members_channel_id_fkey" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "members_inviter_id_fkey" FOREIGN KEY ("inviter_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX "idx_members_channel_id" ON "members" ("channel_id");
CREATE INDEX "idx_members_user_id" ON "members" ("user_id");
CREATE INDEX "idx_members_role" ON "members" ("role");
