CREATE TABLE IF NOT EXISTS public.user_types
(
    id          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name        varchar(255) NOT NULL,
    description text,
    created_at  timestamp(0),
    updated_at  timestamp(0),
    deleted_at  timestamp(0)
);

-- Добавляем базовые типы пользователей
INSERT INTO public.user_types (name, description, created_at)
VALUES 
    ('admin', 'Администратор системы', NOW()),
    ('user', 'Обычный пользователь', NOW()); 