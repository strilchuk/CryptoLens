CREATE TABLE IF NOT EXISTS public.users
(
    id                uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    user_type_id      uuid NOT NULL REFERENCES public.user_types(id),
    nickname          varchar(255) NOT NULL UNIQUE,
    email            varchar(255) NOT NULL UNIQUE,
    password         varchar(255) NOT NULL,
    email_verified_at timestamp(0),
    created_at       timestamp(0),
    updated_at       timestamp(0),
    deleted_at       timestamp(0)
); 