-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
    id serial PRIMARY KEY,
    login varchar(64) UNIQUE NOT NULL CHECK (login SIMILAR TO '[\w\.\-]+'),
    password_hash varchar(256) NOT NULL CHECK (password_hash SIMILAR TO '[\w\.\-]+'),
    public_key text NOT NULL,
    private_key_hash text NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now()
);

CREATE TABLE data (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    name varchar(64) NOT NULL,
    type varchar(64) NOT NULL,
    size_in_bytes integer NOT NULL,
    data bytea NOT NULL,
    external_id text,
    status varchar(64),
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    UNIQUE (user_id, name, type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE data;

DROP TABLE "user";
-- +goose StatementEnd
