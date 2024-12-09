# CNADAssg1
Assg 1 CNAD 

1. Ensure that a .env file is prepared.
2. The .env file should have the appropriate values. 

Format:
PORT=
DB_DSN=
JWT_SECRET_KEY=


Finally, run the main.go file that is in the root folder and the page should be live at the chosen port.

```mermaid
graph TD
    User[User] -->|HTTP Requests| MainService[Main Service]
    MainService -->|API Calls| UserService[User Service]
    MainService -->|API Calls| VehicleService[Vehicle Service]
    MainService -->|API Calls| BillingService[Billing Service]
    UserService -->|CRUD Operations| Database[(Database)]
    VehicleService -->|CRUD Operations| Database[(Database)]
    BillingService -->|CRUD Operations| Database[(Database)]

