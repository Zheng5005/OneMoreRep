-- name: Ping :one
SELECT 1;

-- name: GetRandomQuote :one
SELECT * FROM quote
ORDER BY RANDOM()
LIMIT 1;

-- name: ListQuotes :many
SELECT * FROM quote
ORDER BY author, text;

-- name: GetQuote :one
SELECT * FROM quote WHERE id = $1;

-- name: CreateExercise :one
INSERT INTO exercise (name, target_muscle, notes)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListExercises :many
SELECT * FROM exercise
ORDER BY name;

-- name: GetExercise :one
SELECT * FROM exercise WHERE id = $1;

-- name: UpdateExercise :one
UPDATE exercise
SET name = $2, target_muscle = $3, notes = $4
WHERE id = $1
RETURNING *;

-- name: DeleteExercise :exec
DELETE FROM exercise WHERE id = $1;

-- name: CreateRoutine :one
INSERT INTO routine (name)
VALUES ($1)
RETURNING *;

-- name: ListRoutines :many
SELECT * FROM routine
ORDER BY created_at DESC;

-- name: GetRoutine :one
SELECT * FROM routine WHERE id = $1;

-- name: CreateRoutineExercise :one
INSERT INTO routine_exercise (routine_id, exercise_id, "order")
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListRoutineExercises :many
SELECT re.*, e.name as exercise_name, e.target_muscle
FROM routine_exercise re
JOIN exercise e ON re.exercise_id = e.id
WHERE re.routine_id = $1
ORDER BY re."order";

-- name: CreateWorkoutSession :one
INSERT INTO workout_session (routine_id, started_at)
VALUES ($1, NOW())
RETURNING *;

-- name: EndWorkoutSession :one
UPDATE workout_session
SET ended_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetWorkoutSession :one
SELECT * FROM workout_session WHERE id = $1;

-- name: ListWorkoutSessions :many
SELECT * FROM workout_session
ORDER BY started_at DESC;

-- name: CreateWorkoutSet :one
INSERT INTO workout_set (session_id, exercise_id, set_number, weight, reps)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListWorkoutSetsBySession :many
SELECT ws.*, e.name as exercise_name
FROM workout_set ws
JOIN exercise e ON ws.exercise_id = e.id
WHERE ws.session_id = $1
ORDER BY ws.created_at;

-- name: ListWorkoutSetsByExercise :many
SELECT ws.*, s.started_at as session_started_at
FROM workout_set ws
JOIN workout_session s ON ws.session_id = s.id
WHERE ws.exercise_id = $1
ORDER BY ws.created_at DESC;
