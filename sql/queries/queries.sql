-- name: GetShowSeats :many
SELECT s.id, s.hall_id, s.row_name, s.seat_number 
FROM seats s
WHERE s.hall_id = (SELECT sh.hall_id FROM shows sh WHERE sh.id = $1);

-- name: CheckSeatsAvailableForUpdate :many
-- This locks the rows so no one else can book them while we check
SELECT t.id 
FROM tickets t
WHERE t.show_id = $1 AND t.seat_id = ANY($2::UUID[])
FOR UPDATE;

-- name: CreateBooking :one
INSERT INTO booking (id, show_id, customer_email, customer_phone, status)
VALUES (gen_random_uuid(), $1, $2, $3, 'completed')
RETURNING id, show_id, customer_email, customer_phone, status, created_at;

-- name: CreateTicket :one
INSERT INTO tickets (id, booking_id, show_id, seat_id, qr_code_hash)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
RETURNING id, booking_id, show_id, seat_id, qr_code_hash, created_at;