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

-- name: SearchExercises :many
SELECT * FROM exercise
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%')
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CountExercises :one
SELECT COUNT(*) FROM exercise
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%');

-- name: GetExerciseByNameAndMuscle :one
SELECT * FROM exercise
WHERE name = @name AND target_muscle = @target_muscle
LIMIT 1;

-- name: CountRoutineExercisesByExercise :one
SELECT COUNT(*) FROM routine_exercise WHERE exercise_id = @exercise_id;

-- name: CountWorkoutSetsByExercise :one
SELECT COUNT(*) FROM workout_set WHERE exercise_id = @exercise_id;

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

-- name: UpdateRoutine :one
UPDATE routine
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteRoutine :exec
DELETE FROM routine WHERE id = $1;

-- name: CountRoutines :one
SELECT COUNT(*) FROM routine;

-- name: ListRoutinesPaginated :many
SELECT * FROM routine
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetRoutineExercise :one
SELECT * FROM routine_exercise WHERE id = $1 AND routine_id = $2;

-- name: CountRoutineExercises :one
SELECT COUNT(*) FROM routine_exercise WHERE routine_id = $1;

-- name: GetRoutineExerciseByExercise :one
SELECT * FROM routine_exercise WHERE routine_id = $1 AND exercise_id = $2 LIMIT 1;

-- name: UpdateRoutineExerciseOrder :one
UPDATE routine_exercise
SET "order" = $2
WHERE id = $1
RETURNING *;

-- name: ShiftRoutineExerciseOrderUp :exec
UPDATE routine_exercise
SET "order" = "order" + 1
WHERE routine_id = $1 AND "order" >= $2;

-- name: ShiftRoutineExerciseOrderDown :exec
UPDATE routine_exercise
SET "order" = "order" - 1
WHERE routine_id = $1 AND "order" > $2;

-- name: ReorderRoutineExerciseForward :exec
UPDATE routine_exercise
SET "order" = "order" - 1
WHERE routine_id = $1 AND "order" > $2 AND "order" <= $3;

-- name: ReorderRoutineExerciseBackward :exec
UPDATE routine_exercise
SET "order" = "order" + 1
WHERE routine_id = $1 AND "order" >= $2 AND "order" < $3;

-- name: DeleteRoutineExercise :exec
DELETE FROM routine_exercise WHERE id = $1;

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

-- name: GetActiveWorkoutSession :one
SELECT * FROM workout_session
WHERE ended_at IS NULL
ORDER BY started_at DESC
LIMIT 1;

-- name: GetSessionWithSets :one
SELECT ws.*,
       COALESCE(json_agg(
         json_build_object(
           'id', ws2.id,
           'session_id', ws2.session_id,
           'exercise_id', ws2.exercise_id,
           'set_number', ws2.set_number,
           'weight', ws2.weight,
           'reps', ws2.reps,
           'created_at', ws2.created_at,
           'exercise_name', e.name
         ) ORDER BY ws2.created_at
       ) FILTER (WHERE ws2.id IS NOT NULL), '[]') as sets
FROM workout_session ws
LEFT JOIN workout_set ws2 ON ws.id = ws2.session_id
LEFT JOIN exercise e ON ws2.exercise_id = e.id
WHERE ws.id = $1
GROUP BY ws.id;

-- name: GetWorkoutSet :one
SELECT * FROM workout_set WHERE id = $1;

-- name: UpdateWorkoutSet :one
UPDATE workout_set
SET weight = $2, reps = $3
WHERE id = $1
RETURNING *;

-- name: DeleteWorkoutSet :exec
DELETE FROM workout_set WHERE id = $1;

-- name: RenumberWorkoutSets :exec
UPDATE workout_set
SET set_number = set_number - 1
WHERE session_id = $1 AND exercise_id = $2 AND set_number > $3;

-- name: GetMaxSetNumber :one
SELECT COALESCE(MAX(set_number), 0) FROM workout_set
WHERE session_id = $1 AND exercise_id = $2;

-- name: GetExerciseLastValues :one
SELECT ws.weight, ws.reps
FROM workout_set ws
JOIN workout_session s ON ws.session_id = s.id
WHERE ws.exercise_id = $1 AND s.ended_at IS NOT NULL
ORDER BY s.started_at DESC, ws.created_at DESC
LIMIT 1;

-- name: GetSessionSummary :one
SELECT
  s.id as session_id,
  s.started_at,
  s.ended_at,
  COUNT(DISTINCT ws.exercise_id) as exercise_count,
  COUNT(ws.id) as total_sets,
  COALESCE(SUM(ws.weight * ws.reps), 0) as total_volume
FROM workout_session s
LEFT JOIN workout_set ws ON ws.session_id = s.id
WHERE s.id = $1
GROUP BY s.id, s.started_at, s.ended_at;

-- name: GetSessionExerciseBreakdown :many
SELECT
  ws.exercise_id,
  e.name as exercise_name,
  COUNT(ws.id) as sets_count,
  MAX(ws.weight * ws.reps) as best_volume,
  MAX(ws.weight) as best_weight,
  MAX(ws.reps) as best_reps
FROM workout_set ws
JOIN exercise e ON ws.exercise_id = e.id
WHERE ws.session_id = $1
GROUP BY ws.exercise_id, e.name
ORDER BY e.name;

-- name: GetExerciseHistory :many
SELECT
  ws.id as set_id,
  ws.set_number,
  ws.weight,
  ws.reps,
  ws.created_at as set_created_at,
  s.id as session_id,
  s.started_at as session_started_at,
  s.ended_at as session_ended_at
FROM workout_set ws
JOIN workout_session s ON ws.session_id = s.id
WHERE ws.exercise_id = $1
  AND ($2::text = 'all'
    OR ($2::text = '30d' AND s.started_at >= NOW() - INTERVAL '30 days')
    OR ($2::text = '6m' AND s.started_at >= NOW() - INTERVAL '6 months'))
ORDER BY s.started_at DESC, ws.set_number ASC;

-- name: GetVolumeBySession :many
SELECT
  s.id as session_id,
  s.started_at,
  COALESCE(SUM(ws.weight * ws.reps), 0) as total_volume
FROM workout_session s
LEFT JOIN workout_set ws ON ws.session_id = s.id
WHERE ($1::uuid IS NULL OR ws.exercise_id = $1)
  AND s.ended_at IS NOT NULL
GROUP BY s.id, s.started_at
ORDER BY s.started_at ASC;

-- name: GetVolumeByWeek :many
SELECT
  TO_CHAR(s.started_at, 'YYYY-"W"IW') as period,
  COALESCE(SUM(ws.weight * ws.reps), 0) as total_volume
FROM workout_session s
LEFT JOIN workout_set ws ON ws.session_id = s.id
WHERE ($1::uuid IS NULL OR ws.exercise_id = $1)
  AND s.ended_at IS NOT NULL
GROUP BY TO_CHAR(s.started_at, 'YYYY-"W"IW')
ORDER BY period ASC;

-- name: GetVolumeByMonth :many
SELECT
  TO_CHAR(s.started_at, 'YYYY-MM') as period,
  COALESCE(SUM(ws.weight * ws.reps), 0) as total_volume
FROM workout_session s
LEFT JOIN workout_set ws ON ws.session_id = s.id
WHERE ($1::uuid IS NULL OR ws.exercise_id = $1)
  AND s.ended_at IS NOT NULL
GROUP BY TO_CHAR(s.started_at, 'YYYY-MM')
ORDER BY period ASC;
