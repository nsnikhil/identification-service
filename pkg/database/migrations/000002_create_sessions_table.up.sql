create table if not exists sessions (
    id uuid primary key default gen_random_uuid(),
    userid uuid not null references users (id),
    refreshtoken uuid unique not null,
    revoked boolean not null default false,
    createdat timestamp without time zone default (now() at time zone 'utc'),
    updatedat timestamp without time zone default (now() at time zone 'utc')
);