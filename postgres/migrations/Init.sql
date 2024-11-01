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

ALTER TABLE "game_session" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");
