CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS exercise (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    target_muscle VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS routine (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS routine_exercise (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    routine_id UUID NOT NULL REFERENCES routine(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercise(id) ON DELETE CASCADE,
    "order" INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS workout_session (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    routine_id UUID REFERENCES routine(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS workout_set (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES workout_session(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercise(id) ON DELETE CASCADE,
    set_number INT NOT NULL,
    weight DECIMAL(6,2) NOT NULL DEFAULT 0,
    reps INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS quote (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    text VARCHAR(500) NOT NULL,
    author VARCHAR(255),
    category VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_exercise_name ON exercise(name);
CREATE INDEX IF NOT EXISTS idx_routine_exercise_routine_id ON routine_exercise(routine_id);
CREATE INDEX IF NOT EXISTS idx_workout_set_session_id ON workout_set(session_id);
CREATE INDEX IF NOT EXISTS idx_workout_set_exercise_id ON workout_set(exercise_id);
CREATE INDEX IF NOT EXISTS idx_workout_set_created_at ON workout_set(created_at);
CREATE INDEX IF NOT EXISTS idx_workout_session_started_at ON workout_session(started_at);
