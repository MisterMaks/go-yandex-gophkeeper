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

CREATE TABLE bank_card (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    name varchar(64) NOT NULL,
    number_cipher text NOT NULL,
    expiration_date_cipher text NOT NULL,
    security_code_cipher text NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TABLE login_password (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    name varchar(64) NOT NULL,
    login_cipher text NOT NULL,
    password_cipher text NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TABLE "text" (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    name varchar(64) NOT NULL,
    text text,
    size_in_bytes integer NOT NULL,
    external_id serial,
    status varchar(64),
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TABLE binary_data (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    name varchar(64) NOT NULL,
    data bitea,
    size_in_bytes integer NOT NULL,
    external_id serial,
    status varchar(64),
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE binary_data;

DROP TABLE "text";

DROP TABLE login_password;

DROP TABLE bank_card;

DROP TABLE "user";
-- +goose StatementEnd
