create table if not exists sessions (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users (id),
    refresh_token uuid unique not null,
    revoked boolean not null default false,
    created_at timestamp without time zone default (now() at time zone 'utc'),
    updated_at timestamp without time zone default (now() at time zone 'utc')
);

create index session_user_id_created_at_idx on sessions (user_id, created_at);
create index session_created_at_idx on sessions (created_at);
