## 1. Overview & Objective
A mobile-first workout logging application designed to track gym progress reliably. The core value proposition of the MVP is an offline-first, frictionless logging experience where progress is tracked against independent exercise units rather than specific routines. The architecture is designed for a single-user MVP but structured to scale later into a social platform and marketplace.
## 2. Core Architecture & Data Modeling
To support the requirement that "exercises are independent units," the data model decouples the exercise definition from the routine instance.
### 2.1. Entity Relationship
```
 [Routine] * <------- * [RoutineExercise] * -------> 1 [Exercise]
                                                          |
                                                          | 1
                                                          v
 [WorkoutSession] * <----------------------------------- * [WorkoutSet]
```
### 2.2. Database Schema
This structure ensures that querying historical data for progress graphs only requires looking at the `WorkoutSet` and `Exercise` tables, completely bypassing routine metadata.

**1. Exercise (The Independent Unit)**

|**Field**|**Type**|**Details**|
|---|---|---|
|`id`|UUID|Primary Key|
|`name`|VARCHAR|e.g., "Bench Press - Hypertrophy"|
|`target_muscle`|VARCHAR|e.g., "Chest"|
|`notes`|TEXT|Optional|
|`created_at`|TIMESTAMP||

**2. Routine (The Template)**

|**Field**|**Type**|**Details**|
|---|---|---|
|`id`|UUID|Primary Key|
|`name`|VARCHAR|e.g., "Push Day"|
|`created_at`|TIMESTAMP||
**3. RoutineExercise (Junction Table)**

|**Field**|**Type**|**Details**|
|---|---|---|
|`id`|UUID|Primary Key|
|`routine_id`|UUID|Foreign Key -> Routine|
|`exercise_id`|UUID|Foreign Key -> Exercise|
|`order`|INT|Execution order within the routine|

**4. WorkoutSession (The Active Log)**

|**Field**|**Type**|**Details**|
|---|---|---|
|`id`|UUID|Primary Key|
|`routine_id`|UUID|Foreign Key -> Routine (Optional for freestyles)|
|`started_at`|TIMESTAMP||
|`ended_at`|TIMESTAMP|Optional until workout finishes|

**5. WorkoutSet (The Granular Data)**

|**Field**|**Type**|**Details**|
|---|---|---|
|`id`|UUID|Primary Key|
|`session_id`|UUID|Foreign Key -> WorkoutSession|
|`exercise_id`|UUID|Foreign Key -> Exercise|
|`set_number`|INT|e.g., 1, 2, 3|
|`weight`|DECIMAL||
|`reps`|INT||
|`created_at`|TIMESTAMP||
## 3. MVP Functional Specifications
### 3.1. Routine & Exercise Management
- **Exercise Library:** Users can create custom standalone exercises.
- **Routine Builder:** Users can create weekly routines and assign exercises from the library to these routines using the `RoutineExercise` junction.
- **Reusability:** Adding an existing exercise to a new routine links to the same underlying `exercise_id`, ensuring progress continuity regardless of the active routine.
### 3.2. Active Workout Logging
- **Session Initialization:** Users select a routine to instantiate a `WorkoutSession`.
- **Real-time Logging:** Users input Weight and Reps for each `WorkoutSet`. The UI should default to the values of the last completed session for that specific `exercise_id` to reduce friction.
- **Offline-First Local Storage:** All active logging saves directly to the mobile device's local SQLite database.
### 3.3. Progress Tracking & Visualization
- **Exercise History:** A dedicated view for each exercise showing a historical log of all sets performed.
- **Progress Graphs:** A visual chart plotting Total Volume (Weight × Reps) or 1RM over time for a selected `exercise_id`.
- **Filtering:** Users can view progress across all time, the last 30 days, or the last 6 months.
### 3.4. Motivation & Engagement
- **Post-Workout Screen:** Upon finishing a session (`ended_at` timestamp is set), a modal displays a randomized motivational quote alongside a summary of the workout (Total Volume, Time Elapsed, PRs).
- **Empty State Motivation:** When viewing a graph with no data, or after achieving a milestone, a motivational quote is displayed.
### 3.5. Rest Timer (Secondary Feature)
- **Simple Stopwatch:** A UI element to tap between sets for a pre-defined rest countdown.
- **Background Notification:** Pushes a local alert when time is up.
## 4. Technical Stack & Infrastructure
### 4.1. Client-Side (Mobile)
- **Framework:** React Native (Expo).
- **Local Storage:** WatermelonDB (or similar SQLite wrapper) designed for complex offline-first mapping.
### 4.2 Client-Side (Web)
- **Framework:** React
- **Global state:** Zustand + LocalStorage.
### 4.3. Backend & API
- **Framework:** Golang inside Docker containers.
- **Database:** PostgreSQL (Docker container)
- **Sync Logic (MVP):** A basic REST endpoint handling batched local changes from the client. Since this is a single-user MVP, simple "last write wins" or timestamp-based overwriting is sufficient to resolve state.
### 4.4. Development & Deployment
- **Repository:** Monorepo keeping mobile and backend code in separate directories without complex orchestrators.
- **API Testing:** Bruno collections utilizing environment variables for local vs. production testing.
- **Hosting:** Backend APIs and PostgreSQL database deployed via Render.
## 5. Out of Scope for MVP
- **Complex Conflict Resolution:** CRDTs or operational transformation logic for multi-device concurrent offline sync.
- **AI Integration:** Form checking, automated routine generation, or dynamic weight suggestions.
- **Social Media & Marketplace:** Following users, sharing feeds, selling routines, or hiring coaches.
