create extension if not exists "pgcrypto";

create table if not exists users (
    id uuid primary key default gen_random_uuid(),
    name  varchar(100) not null,
    email  varchar(100) unique not null,
    password_hash  varchar(200) unique not null,
    password_salt  bytea unique not null,
    created_at timestamp without time zone default (now() at time zone 'utc'),
    updated_at timestamp without time zone default (now() at time zone 'utc'),
    check (name <> ''),
    check (email <> ''),
    check (password_hash <> ''),
    check (password_salt <> '')
);