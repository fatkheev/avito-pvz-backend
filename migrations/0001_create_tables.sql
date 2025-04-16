CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица для ПВЗ (Пункт приёма заказов)
CREATE TABLE IF NOT EXISTS pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    city VARCHAR(255) NOT NULL
);

-- Таблица для приёмок товаров
CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    pvz_id UUID NOT NULL REFERENCES pvz(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL
);

-- Таблица для товаров
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    type VARCHAR(50) NOT NULL,
    reception_id UUID NOT NULL REFERENCES receptions(id) ON DELETE CASCADE
);
