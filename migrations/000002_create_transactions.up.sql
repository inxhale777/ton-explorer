CREATE TABLE transactions
(
    hash         text    not null unique,
    account      text    not null,
    success      bool    not null,
    logical_time bigint  not null,
    total_fee    decimal not null,
    comment      text
);