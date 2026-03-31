-- Seed data for seat-booking application
-- Run this after migrations: psql -d cinema_db -f seed.sql

-- Clear existing data (optional - comment out if you want to append)
TRUNCATE tickets, booking, seats, shows, halls, movies RESTART IDENTITY CASCADE;

-- =============================================
-- MOVIES
-- =============================================
INSERT INTO movies (id, title, duration_minutes, description) VALUES
    ('a1b2c3d4-1111-1111-1111-111111111111', 'The Dark Knight', 152, 'When the menace known as the Joker wreaks havoc on Gotham, Batman must face one of the greatest psychological tests.'),
    ('a1b2c3d4-2222-2222-2222-222222222222', 'Inception', 148, 'A thief who steals corporate secrets through dream-sharing technology is given the task of planting an idea.'),
    ('a1b2c3d4-3333-3333-3333-333333333333', 'Interstellar', 169, 'A team of explorers travel through a wormhole in space in an attempt to ensure humanity''s survival.'),
    ('a1b2c3d4-4444-4444-4444-444444444444', 'Dune', 155, 'A noble family becomes embroiled in a war for control over the galaxy''s most valuable asset.'),
    ('a1b2c3d4-5555-5555-5555-555555555555', 'Oppenheimer', 180, 'The story of American scientist J. Robert Oppenheimer and his role in the development of the atomic bomb.');

-- =============================================
-- HALLS
-- =============================================
INSERT INTO halls (id, name, total_seats) VALUES
    ('b1b2c3d4-1111-1111-1111-111111111111', 'Hall A - IMAX', 100),
    ('b1b2c3d4-2222-2222-2222-222222222222', 'Hall B - Premium', 80),
    ('b1b2c3d4-3333-3333-3333-333333333333', 'Hall C - Standard', 120);

-- =============================================
-- SEATS (10 rows x 10 seats per hall for Hall A)
-- =============================================
-- Hall A - IMAX (10x10 = 100 seats)
INSERT INTO seats (id, hall_id, row_name, seat_number) VALUES
    -- Row A
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 1),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 2),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 3),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 4),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 5),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 6),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 7),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 8),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 9),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'A', 10),
    -- Row B
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 1),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 2),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 3),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 4),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 5),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 6),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 7),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 8),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 9),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'B', 10),
    -- Row C
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 1),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 2),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 3),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 4),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 5),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 6),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 7),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 8),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 9),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'C', 10),
    -- Row D
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 1),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 2),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 3),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 4),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 5),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 6),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 7),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 8),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 9),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'D', 10),
    -- Row E
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 1),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 2),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 3),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 4),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 5),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 6),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 7),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 8),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 9),
    (gen_random_uuid(), 'b1b2c3d4-1111-1111-1111-111111111111', 'E', 10);

-- Hall B - Premium (8 rows x 10 seats = 80 seats)
INSERT INTO seats (id, hall_id, row_name, seat_number) VALUES
    -- Row A
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 1),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 2),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 3),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 4),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 5),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 6),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 7),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 8),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 9),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'A', 10),
    -- Row B
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 1),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 2),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 3),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 4),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 5),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 6),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 7),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 8),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 9),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'B', 10),
    -- Row C
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 1),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 2),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 3),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 4),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 5),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 6),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 7),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 8),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 9),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'C', 10),
    -- Row D
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 1),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 2),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 3),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 4),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 5),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 6),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 7),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 8),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 9),
    (gen_random_uuid(), 'b1b2c3d4-2222-2222-2222-222222222222', 'D', 10);

-- =============================================
-- SHOWS (upcoming showtimes)
-- =============================================
INSERT INTO shows (id, movie_id, hall_id, start_time, end_time) VALUES
    -- Today's shows
    ('c1b2c3d4-1111-1111-1111-111111111111', 'a1b2c3d4-1111-1111-1111-111111111111', 'b1b2c3d4-1111-1111-1111-111111111111', 
        CURRENT_DATE + INTERVAL '14 hours', CURRENT_DATE + INTERVAL '16 hours 32 minutes'),
    ('c1b2c3d4-2222-2222-2222-222222222222', 'a1b2c3d4-2222-2222-2222-222222222222', 'b1b2c3d4-2222-2222-2222-222222222222', 
        CURRENT_DATE + INTERVAL '15 hours', CURRENT_DATE + INTERVAL '17 hours 28 minutes'),
    ('c1b2c3d4-3333-3333-3333-333333333333', 'a1b2c3d4-3333-3333-3333-333333333333', 'b1b2c3d4-1111-1111-1111-111111111111', 
        CURRENT_DATE + INTERVAL '18 hours', CURRENT_DATE + INTERVAL '20 hours 49 minutes'),
    ('c1b2c3d4-4444-4444-4444-444444444444', 'a1b2c3d4-1111-1111-1111-111111111111', 'b1b2c3d4-2222-2222-2222-222222222222', 
        CURRENT_DATE + INTERVAL '20 hours', CURRENT_DATE + INTERVAL '22 hours 32 minutes'),
    
    -- Tomorrow's shows
    ('c1b2c3d4-5555-5555-5555-555555555555', 'a1b2c3d4-4444-4444-4444-444444444444', 'b1b2c3d4-1111-1111-1111-111111111111', 
        CURRENT_DATE + INTERVAL '1 day 10 hours', CURRENT_DATE + INTERVAL '1 day 12 hours 35 minutes'),
    ('c1b2c3d4-6666-6666-6666-666666666666', 'a1b2c3d4-5555-5555-5555-555555555555', 'b1b2c3d4-1111-1111-1111-111111111111', 
        CURRENT_DATE + INTERVAL '1 day 14 hours', CURRENT_DATE + INTERVAL '1 day 17 hours'),
    ('c1b2c3d4-7777-7777-7777-777777777777', 'a1b2c3d4-2222-2222-2222-222222222222', 'b1b2c3d4-2222-2222-2222-222222222222', 
        CURRENT_DATE + INTERVAL '1 day 19 hours', CURRENT_DATE + INTERVAL '1 day 21 hours 28 minutes');

-- =============================================
-- Verify seed data
-- =============================================
SELECT 'Movies:' as table_name, COUNT(*) as count FROM movies
UNION ALL
SELECT 'Halls:', COUNT(*) FROM halls
UNION ALL
SELECT 'Seats:', COUNT(*) FROM seats
UNION ALL
SELECT 'Shows:', COUNT(*) FROM shows;
