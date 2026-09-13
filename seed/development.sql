BEGIN;

INSERT INTO coordinators (id, name, email, password_hash)
VALUES
    ('10000000-0000-0000-0000-000000000001', 'Linh Nguyen', 'linh@demo.local', '$2a$10$TasTNzuVtKcoBU5wt4K7ruUL03IRBLMzXjYlkxI.r4ZrpIQ9PYEuC'),
    ('10000000-0000-0000-0000-000000000002', 'Huy Tran', 'huy@demo.local', '$2a$10$TasTNzuVtKcoBU5wt4K7ruUL03IRBLMzXjYlkxI.r4ZrpIQ9PYEuC'),
    ('10000000-0000-0000-0000-000000000003', 'Mai Pham', 'mai@demo.local', '$2a$10$TasTNzuVtKcoBU5wt4K7ruUL03IRBLMzXjYlkxI.r4ZrpIQ9PYEuC')
ON CONFLICT DO NOTHING;

COMMIT;
