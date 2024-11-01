CREATE TABLE IF NOT EXISTS "user" (
  "id" SERIAL PRIMARY KEY,
  "balance" decimal(10,2)
);

CREATE TABLE IF NOT EXISTS  "game_session" (
  "session_id"  SERIAL PRIMARY KEY,
  "user_id" int,
  "bet_amount" decimal(10,2),
  "dice_result" int,
  "won" boolean,
  "active" boolean,
  "session_start" timestamptz,
  "session_end" timestamptz,
  FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION random_decimal(min_val decimal, max_val decimal) 
RETURNS decimal AS $$
BEGIN
    RETURN (random() * (max_val - min_val) + min_val)::decimal(10,2);
END;
$$ LANGUAGE plpgsql;

TRUNCATE TABLE "user" CASCADE;
ALTER SEQUENCE "user_id_seq" RESTART WITH 1;
ALTER SEQUENCE "game_session_session_id_seq" RESTART WITH 1;

WITH generate_series AS (
    SELECT generate_series(1, 1000) AS id
)
INSERT INTO "user" ("balance")
SELECT 
    random_decimal(10, 10000)
FROM 
    generate_series;