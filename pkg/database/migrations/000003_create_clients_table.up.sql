DO $$ BEGIN
    PERFORM 'public.session_strategy'::regtype;
EXCEPTION
    WHEN undefined_object THEN
        create type session_strategy as enum ('revoke_old');
END $$;

create table if not exists clients (
	id uuid primary key default gen_random_uuid(),
	name varchar(100) unique not null,
	secret uuid unique not null default gen_random_uuid(),
	revoked boolean not null default false,
	access_token_ttl integer not null,
	session_ttl integer not null,
	max_active_sessions integer not null,
	private_key bytea unique not null,
	session_strategy session_strategy not null,
	created_at timestamp without time zone default (now() at time zone 'utc'),
	updated_at timestamp without time zone default (now() at time zone 'utc'),
	check (name <> ''),
	check (access_token_ttl >= 1),
	check (session_ttl >= 1),
	check (max_active_sessions >= 1),
	check (session_ttl >= access_token_ttl),
	check (private_key <> '')
);