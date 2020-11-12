create table if not exists clients (
	id uuid primary key default gen_random_uuid(),
	name varchar(100) unique not null,
	secret uuid unique not null default gen_random_uuid(),
	revoked boolean not null default false,
	accesstokenttl integer not null,
	sessionttl integer not null,
	createdat timestamp without time zone default (now() at time zone 'utc'),
	updatedat timestamp without time zone default (now() at time zone 'utc'),
	check(name <> ''),
	check (accesstokenttl > 1),
	check (sessionttl > 1),
	check (sessionttl >= accesstokenttl)
);