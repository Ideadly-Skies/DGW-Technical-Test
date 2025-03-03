# DGW-Technical-Test Requirements

**This is a Revamped DGW Technical Test Repo which hosts the codebase for the following project requirements:**
    
- **Auth API** Feature which *utilizes username and password (hashed)* which will produce a *JWT Token*.
- **Crud Features**: Could be accessed by Someone **Logged-In.**

**Technical Requirement:**
- API must be created using **Go** Programming Language
- Database Utilizes **MySQL / PostgreSQL.**

**Bonus Points (To Be Addressed):**
- Utilizes **clean architecture**
- Utilizes **Dependency Injection** 
- Utilizes **HTTP Framework**
- Provides **API Documentation (Swagger Docs)**
- Create a **Simple UI with web frameworks** such as React/Vue etc. (TBA)

# Features

In this online market place i've implemented a bunch of features with respect to the `admins`, `farmers`, `logs`, `order_items`, `orders`,
`products`, `reviews`, `suppliers`, and `wallet_transactions` entities. The high level functionality overview are as follows:
    
- **product catalog**: farmer could browse through products in the online marketplace which is supplied by the supplier.
- **admin**: the admin is responsible for facilitating the farmers with the transaction which is the logged in the `log` table. The admin has the right to revoke the order if it has passed the stipulated deadline. All the products ordered are logged via the `order_items` linked to the *order ID* of the `order` schema.
- **farmer transaction**: farmer is responsible to pay the outstanding amount for their order at the stipulated deadline. Farmers can opt to choose between two method of transactions: **wallet payment** and **online transaction**
- **review**: after an order have been placed successfully, the farmer could leave a review to which it would be reviewed by a responsible admin. The admin could **Accept** or **Reject** the review made by the farmer.

# Documentation

This project's documentation could be accessed via http://localhost:8080/swagger/index.html. I adhere to the Swaggo framework which is specifically designed for Go application utilizing the Gin web framework.

### Diagrams

- The project's diagram could be accessed via this link: https://excalidraw.com/#json=NOqywD8Gtj12aP56h2Akf,v4bF_iyjctBQCiujX3n17g
- Entity Relationship Diagram could be located in this project root directory slash (/) diagram.