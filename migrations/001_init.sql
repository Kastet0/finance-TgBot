-- Включение расширения для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Сначала создаем таблицу users
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    language_code VARCHAR(10) DEFAULT 'ru',
    currency VARCHAR(3) DEFAULT 'RUB',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Создаем таблицу categories ДО таблицы expenses
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    emoji VARCHAR(10),
    keywords TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Теперь создаем таблицу expenses
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    category VARCHAR(100),
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Создаем индексы
CREATE INDEX IF NOT EXISTS idx_expenses_user_id ON expenses(user_id);
CREATE INDEX IF NOT EXISTS idx_expenses_created_at ON expenses(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);

-- 5. Вставляем предопределенные категории (ДОЛЖНО БЫТЬ ПОСЛЕ СОЗДАНИЯ ТАБЛИЦЫ)
INSERT INTO categories (name, emoji, keywords) VALUES
    ('Еда', '🍔', ARRAY['кафе', 'ресторан', 'макдональдс', 'бургер', 'пицца', 'суши', 'кофе']),
    ('Продукты', '🛒', ARRAY['пятерочка', 'магнит', 'ашан', 'лента', 'продукты']),
    ('Транспорт', '🚗', ARRAY['такси', 'метро', 'автобус', 'бензин', 'заправка']),
    ('Коммуналка', '🏠', ARRAY['квартплата', 'электричество', 'вода', 'газ', 'интернет']),
    ('Развлечения', '🎬', ARRAY['кино', 'концерт', 'театр', 'музей', 'парк']),
    ('Здоровье', '🏥', ARRAY['аптека', 'врач', 'больница', 'лекарства', 'спортзал']),
    ('Одежда', '👕', ARRAY['одежда', 'обувь', 'магазин', 'шопинг']),
    ('Образование', '📚', ARRAY['курсы', 'книги', 'учебник', 'обучение']),
    ('Подарки', '🎁', ARRAY['подарок', 'сюрприз', 'день рождения']),
    ('Прочее', '📦', ARRAY['другое', 'прочее'])
ON CONFLICT (name) DO NOTHING;

-- 6. Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 7. Создаем триггеры
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_expenses_updated_at 
    BEFORE UPDATE ON expenses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();