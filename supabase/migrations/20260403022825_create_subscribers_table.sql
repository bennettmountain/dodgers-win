CREATE TABLE IF NOT EXISTS subscribers (
    id int primary key generated always as identity,
    phone_number varchar(15) unique not null,
    unsubscribed boolean default false,
    created timestamp default current_timestamp
);
