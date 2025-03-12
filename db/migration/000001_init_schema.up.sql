CREATE TABLE "users" (
  "user_id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realms" (
  "realm_id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "owner_nickname" varchar NOT NULL,
  "owner_rm_id" bigint UNIQUE NOT NULL,
  "capitals" int[],
  "political_entity" varchar NOT NULL,
  "color" varchar NOT NULL,
  "population_growth_rate" float NOT NULL,
  "state_coffers" int NOT NULL,
  "census_at" timestamptz NOT NULL,
  "tax_collection_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realm_members" (
  "rm_id" bigint PRIMARY KEY,
  "realm_id" bigint,
  "nickname" varchar UNIQUE NOT NULL,
  "status" varchar NOT NULL,
  "private_coffers" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "member_authorities" (
  "rm_id" bigint UNIQUE NOT NULL,
  "create_unit" bool NOT NULL,
  "reinforce_unit" bool NOT NULL,
  "move_unit" bool NOT NULL,
  "attack_unit" bool NOT NULL,
  "private_troops" bool NOT NULL,
  "census" bool NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sectors" (
  "cell_number" int PRIMARY KEY,
  "province_number" int NOT NULL,
  "realm_id" bigint NOT NULL,
  "rm_id" bigint NOT NULL,
  "population" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realm_sectors_jsonb" (
  "realm_sectors_jsonb_id" bigint UNIQUE NOT NULL,
  "cells_jsonb" JSONB NOT NULL
);

CREATE TABLE "levies" (
  "levy_id" bigserial PRIMARY KEY,
  "stationed" bool NOT NULL,
  "name" varchar NOT NULL,
  "morale" smallint NOT NULL,
  "encampment" int NOT NULL,
  "swordmen" int NOT NULL,
  "shield_bearers" int NOT NULL,
  "archers" int NOT NULL,
  "lancers" int NOT NULL,
  "supply_troop" int NOT NULL,
  "movement_speed" float NOT NULL,
  "rm_id" bigint NOT NULL,
  "realm_id" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "levies_actions" (
  "levy_action_id" bigserial PRIMARY KEY,
  "levy_id" bigint NOT NULL,
  "realm_id" bigint NOT NULL,
  "rm_id" bigint NOT NULL,
  "origin_sector" int NOT NULL,
  "target_sector" int NOT NULL,
  "distance" float NOT NULL,
  "action_type" varchar NOT NULL,
  "completed" boolean NOT NULL,
  "target_realm_id" bigint,
  "started_at" timestamptz NOT NULL,
  "expected_completion_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "indigenous_units" (
  "sector_number" int PRIMARY KEY,
  "swordmen" int NOT NULL,
  "archers" int NOT NULL,
  "lancers" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "conquered_nations" (
  "conquered_nation_id" bigserial PRIMARY KEY,
  "rm_id" bigint NOT NULL,
  "owner_nickname" varchar NOT NULL,
  "country_name" varchar NOT NULL,
  "cells_jsonb" JSONB NOT NULL,
  "conquered_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "battle_outcomes" (
  "levy_action_id" bigint NOT NULL,
  "realm_id" bigint NOT NULL,
  "attacker" JSONB NOT NULL,
  "defender" JSONB NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "levy_surrenders" (
  "surrender_reason" varchar NOT NULL,
  "surrendered_at" timestamptz NOT NULL,
  "surrendered_sector_location" int NOT NULL,
  "levy_id" bigint UNIQUE NOT NULL,
  "receiving_realm_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "world_time_records" (
  "world_time_record_id" bigserial PRIMARY KEY,
  "stop_reason" varchar NOT NULL,
  "world_stopped_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "state_coffers_logs" (
  "state_coffers_log_id" bigserial PRIMARY KEY,
  "realm_id" bigint NOT NULL,
  "change_amount" int NOT NULL,
  "total_coffers" int NOT NULL,
  "reason" varchar NOT NULL,
  "world_time_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "private_coffers_logs" (
  "private_coffers_log_id" bigserial PRIMARY KEY,
  "rm_id" bigint NOT NULL,
  "change_amount" int NOT NULL,
  "total_coffers" int NOT NULL,
  "reason" varchar NOT NULL,
  "world_time_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("email");

CREATE INDEX ON "realms" ("owner_rm_id");

CREATE INDEX ON "realm_members" ("realm_id");

CREATE INDEX ON "member_authorities" ("rm_id");

CREATE INDEX ON "sectors" ("realm_id");

CREATE INDEX ON "sectors" ("rm_id");

CREATE INDEX ON "realm_sectors_jsonb" ("realm_sectors_jsonb_id");

CREATE INDEX ON "levies" ("rm_id");

CREATE INDEX ON "levies" ("realm_id");

CREATE INDEX ON "levies" ("encampment");

CREATE INDEX ON "levies" ("stationed");

CREATE INDEX ON "levies_actions" ("rm_id");

CREATE INDEX ON "levies_actions" ("levy_id");

CREATE INDEX ON "levies_actions" ("realm_id");

CREATE INDEX ON "levies_actions" ("expected_completion_at");

CREATE INDEX ON "levies_actions" ("completed");

CREATE INDEX ON "conquered_nations" ("rm_id");

CREATE INDEX ON "battle_outcomes" ("levy_action_id");

CREATE INDEX ON "battle_outcomes" ("realm_id");

CREATE INDEX ON "battle_outcomes" ("created_at");

CREATE INDEX ON "levy_surrenders" ("levy_id");

CREATE INDEX ON "levy_surrenders" ("receiving_realm_id");

CREATE INDEX ON "world_time_records" ("created_at");

CREATE INDEX ON "state_coffers_logs" ("created_at");

CREATE INDEX ON "private_coffers_logs" ("created_at");

ALTER TABLE "realm_members" ADD FOREIGN KEY ("rm_id") REFERENCES "users" ("user_id") ON DELETE SET NULL;

ALTER TABLE "realm_members" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "member_authorities" ADD FOREIGN KEY ("rm_id") REFERENCES "realm_members" ("rm_id");

ALTER TABLE "sectors" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "sectors" ADD FOREIGN KEY ("rm_id") REFERENCES "realm_members" ("rm_id");

ALTER TABLE "realm_sectors_jsonb" ADD FOREIGN KEY ("realm_sectors_jsonb_id") REFERENCES "realms" ("realm_id") ON DELETE CASCADE;

ALTER TABLE "levies" ADD FOREIGN KEY ("encampment") REFERENCES "sectors" ("cell_number");

ALTER TABLE "levies" ADD FOREIGN KEY ("rm_id") REFERENCES "realm_members" ("rm_id");

ALTER TABLE "levies" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id") ON DELETE SET NULL;

ALTER TABLE "levies_actions" ADD FOREIGN KEY ("levy_id") REFERENCES "levies" ("levy_id") ON DELETE CASCADE;

ALTER TABLE "levies_actions" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id") ON DELETE SET NULL;

ALTER TABLE "levies_actions" ADD FOREIGN KEY ("rm_id") REFERENCES "realm_members" ("rm_id");

ALTER TABLE "conquered_nations" ADD FOREIGN KEY ("rm_id") REFERENCES "realm_members" ("rm_id");

ALTER TABLE "battle_outcomes" ADD FOREIGN KEY ("levy_action_id") REFERENCES "levies_actions" ("levy_action_id") ON DELETE CASCADE;

ALTER TABLE "battle_outcomes" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "levy_surrenders" ADD FOREIGN KEY ("levy_id") REFERENCES "levies" ("levy_id") ON DELETE CASCADE;

ALTER TABLE "levy_surrenders" ADD FOREIGN KEY ("receiving_realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "state_coffers_logs" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id") ON DELETE CASCADE;

ALTER TABLE "private_coffers_logs" ADD FOREIGN KEY ("rm_id") REFERENCES "realm_members" ("rm_id") ON DELETE CASCADE;
