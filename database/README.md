Database Design Overview

Database Name: ecs_system

---

1. Users Table
Purpose:  
The `users` table stores all information related to user accounts, including authentication and membership tiers. This table ensures secure user data handling and supports the differentiation of membership benefits.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `user_id`         | INT (PK, AI)      | Unique identifier for each user.      |
| `email`           | VARCHAR(255)      | User's email address (unique).        |
| `password_hash`   | VARCHAR(255)      | Encrypted password for authentication.|
| `membership_tier` | ENUM('Basic', 'Premium', 'VIP') | Membership type.  |
| `phone_number`    | VARCHAR(15)       | User's phone number (optional).       |
| `country_code`    | VARCHAR(5)        | Country code for the user's phone number. |
| `created_at`      | DATETIME          | Timestamp of registration.            |
| `updated_at`      | DATETIME          | Timestamp of last profile update.     |

Why?
- To securely store user data with minimal redundancy.
- To manage different membership tiers for billing and feature access.

---

2. Vehicles Table
Purpose:  
The `vehicles` table holds information about electric vehicles available for reservation. It tracks vehicle status, location, and battery charge level to ensure accurate availability and readiness.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `vehicle_id`      | INT (PK, AI)      | Unique identifier for each vehicle.   |
| `model`           | VARCHAR(255)      | Vehicle model name.                   |
| `license_plate`   | VARCHAR(50)       | Vehicle license plate (unique).       |
| `charge_level`    | INT               | Battery charge level (percentage).    |
| `availability`    | BOOLEAN           | Vehicle availability status.          |
| `location`        | VARCHAR(255)      | Current location of the vehicle.      |
| `created_at`      | DATETIME          | Timestamp of vehicle addition.        |

Why? 
- To track vehicle availability and readiness in real-time.
- To ensure separation of vehicle data from reservation details for modularity.

---

3. Reservations Table
Purpose: 
The `reservations` table logs all bookings made by users, including the reservation period, status, and associated user and vehicle IDs. This enables accurate tracking of booking histories.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `reservation_id`  | INT (PK, AI)      | Unique identifier for the reservation.|
| `user_id`         | INT (FK)          | Refers to the `users` table.          |
| `vehicle_id`      | INT (FK)          | Refers to the `vehicles` table.       |
| `start_time`      | DATETIME          | Reservation start time.               |
| `end_time`        | DATETIME          | Reservation end time.                 |
| `status`          | ENUM('Booked', 'Cancelled', 'Completed') | Reservation status. |
| `created_at`      | DATETIME          | Timestamp of reservation creation.    |

Why? 
- To track the lifecycle of reservations for billing and analytics.
- To normalize relationships between users and vehicles.

---

4. Billing Table
Purpose: 
The `billing` table tracks payment details and invoice generation for reservations. It is linked to the `reservations` table to ensure accurate cost calculations and refunds.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `billing_id`      | INT (PK, AI)      | Unique identifier for the billing entry. |
| `reservation_id`  | INT (FK)          | Refers to the `reservations` table.    |
| `amount`          | DECIMAL(10, 2)    | Total amount charged.                 |
| `payment_status`  | ENUM('Pending', 'Paid', 'Refunded') | Payment status.   |
| `created_at`      | DATETIME          | Timestamp of billing creation.        |

Why?
- To manage financial transactions and billing histories.
- To centralize payment information for refunds and analytics.

---

5. Promotions Table
Purpose: 
The `promotions` table stores details about active promotional discounts. It allows flexible management of discounts and ensures separation from the billing logic.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `promotion_id`    | INT (PK, AI)      | Unique identifier for the promotion.  |
| `name`            | VARCHAR(255)      | Name of the promotion.                |
| `discount_percent`| DECIMAL(5, 2)     | Discount percentage.                  |
| `valid_from`      | DATETIME          | Start date of the promotion.          |
| `valid_to`        | DATETIME          | End date of the promotion.            |

Why? 
- To allow dynamic application of discounts without modifying core billing logic.
- To track promotional periods for reporting.

---

6. User Promotions Table
Purpose:  
The `user_promotions` table links users to applied promotions. It ensures flexibility and avoids duplicating promotion information in the `users` or `billing` tables.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `id`              | INT (PK, AI)      | Unique identifier.                    |
| `user_id`         | INT (FK)          | Refers to the `users` table.          |
| `promotion_id`    | INT (FK)          | Refers to the `promotions` table.     |
| `applied_at`      | DATETIME          | Timestamp of application.             |

Why? 
- To keep track of promotions applied to specific users.
- To reduce redundancy by separating promotional data.

---

7. Rental History Table
Purpose:
The `rental_history` table records user rentals for tracking history and generating detailed reports.

| Column Name       | Data Type         | Description                           |
|-------------------|-------------------|---------------------------------------|
| `id`              | INT (PK, AI)      | Unique identifier.                    |
| `user_id`         | INT (FK)          | Refers to the `users` table.          |
| `vehicle_id`      | INT (FK)          | Refers to the `vehicles` table.       |
| `start_time`      | DATETIME          | Start time of the rental.             |
| `end_time`        | DATETIME          | End time of the rental.               |
| `total_cost`      | DECIMAL(10,2)     | Total cost of the rental.             |

Why?
- Provide a clear record of user rental history.
- Enable reporting and profile management features.

---

Relationships and Normalization
1. 1NF (Atomic Columns): Each table stores atomic data, avoiding arrays or multiple values in a single column.
2. 2NF (Full Dependency): All non-primary-key columns depend entirely on the primary key.
3. 3NF (No Transitive Dependency): No column depends on another non-primary column (e.g., `email` is stored only in `users`).

---

Entity-Relationship Summary
- Users ↔ Reservations: One user can have multiple reservations. (1-to-Many)
- Vehicles ↔ Reservations: One vehicle can have multiple reservations. (1-to-Many)
- Reservations ↔ Billing: One reservation corresponds to one billing entry. (1-to-1)
- Users ↔ User Promotions: One user can use multiple promotions. (1-to-Many)
- Promotions ↔ User Promotions: One promotion can apply to multiple users. (1-to-Many)
- User ↔ Rental History: One promotion can have multiple rental records. (1-to-Many)

This design ensures modularity, scalability, and ease of maintenance for the electric car-sharing system.