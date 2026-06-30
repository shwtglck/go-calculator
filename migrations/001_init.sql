-- Таблица для хранения истории вычислений.
-- Каждая строка = одна операция, которую сделал сервис.
CREATE TABLE IF NOT EXISTS calculations (
    id         SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    operand_a  DOUBLE PRECISION NOT NULL,
    operand_b  DOUBLE PRECISION NOT NULL,
    operator   VARCHAR(10) NOT NULL,
    result     DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
