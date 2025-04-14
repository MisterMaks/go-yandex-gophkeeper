-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
    id serial PRIMARY KEY,
    login varchar(64) UNIQUE NOT NULL CHECK (login SIMILAR TO '[\w\.\-]+'),
    password_hash bytea NOT NULL,
    public_key bytea NOT NULL,
    private_key_cipher bytea NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE data (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    name varchar(64) NOT NULL,
    type varchar(64) NOT NULL CHECK (type IN ('LOGIN_PASSWORD', 'BANK_CARD', 'TEXT', 'BINARY')),
    data bytea,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    UNIQUE (user_id, name, type)  -- unique name for type and user
);

CREATE INDEX name_index ON data (name);
CREATE INDEX type_index ON data (type);
CREATE INDEX created_at_index ON data (created_at);
CREATE INDEX updated_at_index ON data (updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE data;

DROP TABLE "user";
-- +goose StatementEnd
