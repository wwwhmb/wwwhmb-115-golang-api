create extension pgcrypto;

create table users (
    user_id bigserial primary key,
    username text not null unique check(length(username) >= 3),
    password text not null,
    created_at timestamp not null default now()
);

create table posts (
    post_id bigserial primary key,
    user_id bigint not null references users(user_id),
    message text not null,
    created_at timestamp not null default now()
);
create index on posts(user_id, created_at);

create table sessions (
    session_id uuid primary key default gen_random_uuid(),
    user_id bigint not null references users(user_id),
    created_at timestamp not null default now()
);
