create table if not exists sessions (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users (id),
    refresh_token uuid unique not null,
    revoked boolean not null default false,
    created_at timestamp without time zone default (now() at time zone 'utc'),
    updated_at timestamp without time zone default (now() at time zone 'utc')
);