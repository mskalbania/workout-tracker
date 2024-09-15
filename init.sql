CREATE TABLE "user"
(
    id            uuid PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL
);