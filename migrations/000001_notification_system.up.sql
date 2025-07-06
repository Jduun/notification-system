create extension if not exists "uuid-ossp";

create table notifications (
    id uuid primary key default uuid_generate_v4(),
    delivery_type text not null,
    recipient text not null,
    content text not null,
    status text not null default 'pending',
    retries smallint not null default 0,
    created_at timestamp not null default now(),
    sent_at timestamp,
    check (delivery_type in ('email', 'sms', 'telegram')),
    check (status in ('delivered', 'pending', 'in_queue', 'retrying', 'failed'))
);