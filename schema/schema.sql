
CREATE TYPE user_role AS ENUM ('admin', 'user', 'guest');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'pending', 'deleted');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    status user_status NOT NULL,
    last_login_at BIGINT DEFAULT NULL,
    deleted_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_name ON users(first_name, last_name);
CREATE INDEX idx_users_full_name ON users(full_name);

CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    token TEXT NOT NULL,
    revoked_at BIGINT DEFAULT NULL,
    expires_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_tokens_token ON tokens(token);


CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_url VARCHAR(255) NOT NULL UNIQUE,
    cover_url TEXT,
    name VARCHAR(255) NOT NULL,
    channel_type VARCHAR(50) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    last_message_id UUID,
    last_read_message_id UUID,
    is_public BOOLEAN NOT NULL DEFAULT false,
    is_distinct BOOLEAN NOT NULL DEFAULT false,
    create_by UUID NOT NULL REFERENCES users(id),
    create_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    update_by UUID NOT NULL REFERENCES users(id),
    update_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

CREATE INDEX idx_channels_channel_url ON channels(channel_url);
CREATE INDEX idx_channels_channel_type ON channels(channel_type);
CREATE INDEX idx_channels_is_public ON channels(is_public);
CREATE INDEX idx_channels_create_by ON channels(create_by);
CREATE INDEX idx_channels_name ON channels(name);
CREATE INDEX idx_channels_is_distinct ON channels(is_distinct);

CREATE TABLE members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    state VARCHAR(40) NOT NULL DEFAULT 'joined',
    joined_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    inviter_id UUID REFERENCES users(id),
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    CONSTRAINT members_channel_user_key UNIQUE (channel_id, user_id)
);

CREATE INDEX idx_members_channel_id ON members(channel_id);
CREATE INDEX idx_members_user_id ON members(user_id);
CREATE INDEX idx_members_role ON members(role);
CREATE INDEX idx_members_state ON members(state);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id),
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    message_type VARCHAR(50) NOT NULL DEFAULT 'MESG',
    custom_type VARCHAR(50) NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}',
    req_id VARCHAR(255),
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

CREATE INDEX idx_messages_channel_id_created_at ON messages(channel_id, created_at DESC);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);