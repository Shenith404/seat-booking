-- name: GetShowSeats :many
SELECT id, hall_id, row_name, seat_number 
FROM seats 
WHERE hall_id = (SELECT hall_id FROM shows WHERE id = $1);

-- name: CheckSeatsAvailableForUpdate :many
-- This locks the rows so no one else can book them while we check
SELECT id 
FROM tickets 
WHERE show_id = $1 AND seat_id = ANY($2::UUID[])
FOR UPDATE;

-- name: CreateBooking :one
INSERT INTO bookings (show_id, customer_email, customer_phone, status)
VALUES ($1, $2, $3, 'completed')
RETURNING booking_id,customer_email,customer_phone,status,created_at;

-- name: CreateTicket :one
INSERT INTO tickets (booking_id, show_id, seat_id, qr_code_hash)
VALUES ($1, $2, $3, $4)
RETURNING id;