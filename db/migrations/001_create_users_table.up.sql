create table if not exists users
(
    id bigserial primary key,
    login varchar(255) not null
        constraint users_login_unique unique,
    password varchar(255) not null,
    created_at timestamp(0) not null,
    updated_at timestamp(0) not null
);