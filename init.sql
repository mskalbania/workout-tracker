CREATE TABLE "user"
(
    id            uuid PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE exercise
(
    id           uuid PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    description  TEXT         NOT NULL,
    category     VARCHAR(255) NOT NULL,
    muscle_group VARCHAR(255) NOT NULL
);

CREATE TABLE workout
(
    id      uuid PRIMARY KEY,
    name    VARCHAR(255) NOT NULL,
    "owner" uuid NOT NULL,
    comment TEXT
);

CREATE INDEX workout_owner_index ON workout ("owner");

CREATE TABLE workout_exercise
(
    workout_exercise_id uuid PRIMARY KEY,
    workout_id          uuid NOT NULL REFERENCES workout (id) ON DELETE CASCADE,
    exercise_id         uuid NOT NULL REFERENCES exercise (id),
    "order"             int  NOT NULL,
    repetitions         int  NOT NULL,
    sets                int  NOT NULL,
    weight  DECIMAL(5, 2),
    comment TEXT
);

CREATE INDEX workout_exercise_workout_id_index ON workout_exercise (workout_id);
CREATE INDEX workout_exercise_exercise_id_index ON workout_exercise (exercise_id);

CREATE TABLE workout_schedule
(
    id          uuid PRIMARY KEY,
    "owner"     uuid      NOT NULL,
    plan        uuid      NOT NULL REFERENCES workout (id) On DELETE CASCADE,
    scheduledAt TIMESTAMP NOT NULL
);