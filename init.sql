CREATE TABLE "user"
(
    id            uuid PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
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
    id           uuid PRIMARY KEY,
    "owner"      uuid      NOT NULL,
    workout      uuid      NOT NULL REFERENCES workout (id) On DELETE CASCADE,
    scheduled_at TIMESTAMP NOT NULL,
    crated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AT TIME ZONE 'UTC',
    completed    boolean   NOT NULL DEFAULT FALSE
);

-- Populating here since this is some predefined data
INSERT INTO exercise (id, name, description, category, muscle_group)
VALUES ('87df312d-36e0-40e8-915e-093ac3342ac8', 'Bench Press',
        'The bench press is an upper-body weight training exercise.', 'STRENGTH', 'CHEST'),
       ('c3339fa8-f9d6-481d-b983-f9cdc24ca4d0', 'Squat',
        'The squat is a lower body exercise.', 'STRENGTH', 'LEGS'),
       ('94b4109b-25ba-4519-8aa7-6adef75c0d37', 'Pull-up',
        'A pull-up is an upper-body strength exercise.', 'CARDIO', 'BACK'),
       ('66a27a50-191d-4338-a6b9-59366b9c423c', 'Push-up',
        'A push-up is a common calisthenics exercise beginning from the prone position.', 'STRENGTH', 'CHEST');
