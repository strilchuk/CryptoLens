CREATE TABLE IF NOT EXISTS public.bybit_accounts
(
    id           bigserial
        PRIMARY KEY,
    user_id      uuid NOT NULL REFERENCES public.users(id),
    api_key      varchar(255)                                      NOT NULL,
    api_secret   varchar(255)                                      NOT NULL,
    account_type varchar(255) DEFAULT 'UNIFIED'::character varying NOT NULL,
    is_active    boolean      DEFAULT true                         NOT NULL,
    created_at   timestamp(0),
    updated_at   timestamp(0),
    deleted_at   timestamp(0)
); 