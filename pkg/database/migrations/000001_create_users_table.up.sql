create extension if not exists "pgcrypto";

create table if not exists users (
    id uuid primary key default gen_random_uuid(),
    name  varchar(100) not null,
    email  varchar(100) unique not null,
    passwordhash  varchar(200) unique not null,
    passwordsalt  bytea unique not null,
    createdat timestamp without time zone default (now() at time zone 'utc'),
    updatedat timestamp without time zone default (now() at time zone 'utc'),
    check (name <> ''),
    check (email <> ''),
    check (passwordhash <> ''),
    check (passwordsalt <> '')
);