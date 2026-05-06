-- +goose Up

-- properties — JSONB-колонка для типоспецифичных свойств детали.
-- У каждого типа свой набор полей: у корпуса — strength, у двигателя — class и required_strength,
-- у щита — shield_type, у оружия — weapon_type. JSONB позволяет хранить их в одной таблице
-- без создания отдельных таблиц на каждый тип.
-- DEFAULT '{}' — новые детали без свойств получают пустой объект, а не NULL.
--
-- reserved — сколько единиц этой детали уже зарезервировано под заказы.
-- Доступно для новых заказов: stock_quantity - reserved.
ALTER TABLE parts
    ADD COLUMN properties JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN reserved INT NOT NULL DEFAULT 0;

-- CHECK-ограничения — защита на уровне БД от невалидных состояний.
-- Даже если в коде баг, PostgreSQL не даст записать отрицательный остаток
-- или зарезервировать больше, чем есть на складе.
--
-- chk_stock_non_negative  — stock_quantity >= 0 (нельзя уйти в минус по складу)
-- chk_reserved_non_negative — reserved >= 0 (нельзя «отменить» больше, чем зарезервировано)
-- chk_reserved_le_stock   — reserved <= stock_quantity (нельзя зарезервировать больше, чем есть)
ALTER TABLE parts
    ADD CONSTRAINT chk_stock_non_negative CHECK (stock_quantity >= 0),
    ADD CONSTRAINT chk_reserved_non_negative CHECK (reserved >= 0),
    ADD CONSTRAINT chk_reserved_le_stock CHECK (reserved <= stock_quantity);

-- +goose Down

-- Удаляем в обратном порядке: сначала ограничения, потом колонки.
-- IF EXISTS — чтобы откат не падал, если миграцию применяли частично.
ALTER TABLE parts
DROP CONSTRAINT IF EXISTS chk_reserved_le_stock,
    DROP CONSTRAINT IF EXISTS chk_reserved_non_negative,
    DROP CONSTRAINT IF EXISTS chk_stock_non_negative;

ALTER TABLE parts
DROP COLUMN IF EXISTS reserved,
    DROP COLUMN IF EXISTS properties;