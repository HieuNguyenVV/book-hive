-- Add distinct flag for 1:1 channels
ALTER TABLE "channels" ADD COLUMN "is_distinct" boolean NOT NULL DEFAULT false;

CREATE INDEX "idx_channels_is_distinct" ON "channels" ("is_distinct");

-- Messages table
CREATE TABLE "messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "channel_id" uuid NOT NULL,
  "sender_id" uuid NOT NULL,
  "content" text NOT NULL,
  "message_type" character varying(50) NOT NULL DEFAULT 'MESG',
  "custom_type" character varying(50) NOT NULL DEFAULT '',
  "data" jsonb NOT NULL DEFAULT '{}',
  "req_id" character varying(255) NULL,
  "created_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  "updated_at" bigint NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::bigint,
  PRIMARY KEY ("id"),
  CONSTRAINT "messages_channel_id_fkey" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messages_sender_id_fkey" FOREIGN KEY ("sender_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX "idx_messages_channel_id_created_at" ON "messages" ("channel_id", "created_at" DESC);
CREATE INDEX "idx_messages_sender_id" ON "messages" ("sender_id");

-- Member state for join tracking
ALTER TABLE "members" ADD COLUMN "state" character varying(40) NOT NULL DEFAULT 'joined';

CREATE INDEX "idx_members_state" ON "members" ("state");
