-- Insert data into users
INSERT INTO users (email, password_hash, membership_tier, phone_number, country_code, phone_verified, created_at)
VALUES
('haziqrazak14.27@gmail.com', '$2a$10$OjP5n2meN4Gj1sCgzOBdcegCQth/nshUNu/xsnmi0mEm/s3bQ2Rwm', 'VIP', '83048488', '+65', TRUE, NOW()),
('alice@example.com', '$2a$10$DkF1f2m9Khz9yJHd2MqYuOMbZjm5wyPryOaR/u50AVzAeQ3Dj43G6', 'Premium', '512345678', '+1', TRUE, NOW()),
('bob@example.com', '$2a$10$6g8Tz1o5QUH.Qx4EpREhEu/HFsX8hbZ8oM2PEHzH1B6zGUN4dJ.gK', 'Basic', '441234567', '+44', TRUE, NOW());

-- Insert data into vehicles
INSERT INTO vehicles (model, license_plate, charge_level, cleanliness, last_service_date, availability, location)
VALUES
('Tesla Model 3', 'SGP1234T', 80, 'Clean', '2024-03-01 12:00:00', TRUE, 'Orchard Road, Singapore'),
('Hyundai Kona', 'SGP5678U', 100, 'Needs Cleaning', '2024-02-28 15:30:00', TRUE, 'Jurong East, Singapore'),
('Nissan Leaf', 'SGP9012X', 50, 'Clean', '2024-01-15 09:00:00', FALSE, 'Changi Airport, Singapore');

-- Insert data into promotions
INSERT INTO promotions (name, discount_percent, valid_from, valid_to)
VALUES
('New Year Discount', 15.00, '2024-01-01', '2024-01-31'),
('Loyalty Reward', 10.00, '2024-02-01', '2024-12-31');

-- Insert data into user promotions
INSERT INTO user_promotions (user_id, promotion_id, applied_at)
VALUES
(1, 1, '2024-01-02 10:00:00'),
(2, 2, '2024-02-15 14:30:00');

-- Insert data into reservations
INSERT INTO reservations (user_id, vehicle_id, start_time, end_time, status, created_at)
VALUES
(1, 1, '2024-03-01 09:00:00', '2024-03-01 13:00:00', 'Completed', NOW()),
(2, 2, '2024-03-02 10:00:00', '2024-03-02 15:00:00', 'Cancelled', NOW()),
(3, 3, '2024-03-03 08:00:00', '2024-03-03 11:00:00', 'Booked', NOW());

-- Insert data into billing
INSERT INTO billing (reservation_id, amount, payment_status, invoice_id, created_at)
VALUES
(1, 120.00, 'Paid', 'INV-20240301-1', NOW()),
(2, 0.00, 'Refunded', 'INV-20240302-2', NOW()),
(3, 75.00, 'Pending', 'INV-20240303-3', NOW());

-- Insert data into rental history
INSERT INTO rental_history (user_id, vehicle_id, start_time, end_time, total_cost)
VALUES
(1, 1, '2024-03-01 09:00:00', '2024-03-01 13:00:00', 120.00),
(2, 2, '2024-03-02 10:00:00', '2024-03-02 15:00:00', 0.00),
(3, 3, '2024-03-03 08:00:00', '2024-03-03 11:00:00', 75.00);
