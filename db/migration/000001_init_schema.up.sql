CREATE TABLE "users" (
  "user_id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "nickname" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realms" (
  "realm_id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "owner_nickname" varchar NOT NULL,
  "owner_id" bigint UNIQUE NOT NULL,
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
  "realm_member_id" bigserial PRIMARY KEY,
  "user_id" bigint UNIQUE NOT NULL,
  "status" varchar NOT NULL,
  "private_money" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sectors" (
  "cell_number" int PRIMARY KEY,
  "province_number" int NOT NULL,
  "realm_id" bigint NOT NULL,
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
  "realm_member_id" bigint NOT NULL,
  "realm_id" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "levies_actions" (
  "levy_action_id" bigserial PRIMARY KEY,
  "levy_id" bigint NOT NULL,
  "origin_sector" int NOT NULL,
  "target_sector" int NOT NULL,
  "action_type" varchar NOT NULL,
  "completed" boolean NOT NULL,
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
  "owner_id" bigint NOT NULL,
  "owner_nickname" varchar NOT NULL,
  "country_name" varchar NOT NULL,
  "cells_jsonb" JSONB NOT NULL,
  "conquered_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "battle_outcome" (
  "levy_action_id" bigint NOT NULL,
  "swordman_casualties" int NOT NULL,
  "archer_casualties" int NOT NULL,
  "shield_bearer_casualties" int NOT NULL,
  "lancer_casualties" int NOT NULL,
  "supply_troop_casualties" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "levy_surrenders" (
  "surrender_reason" varchar NOT NULL,
  "surrendered_at" timestamptz NOT NULL,
  "surrendered_sector_location" int NOT NULL,
  "levy_id" bigint UNIQUE NOT NULL,
  "receiving_realm_id" bigint NOT NULL
);

CREATE INDEX ON "users" ("email");

CREATE INDEX ON "realms" ("owner_id");

CREATE INDEX ON "realm_members" ("user_id");

CREATE INDEX ON "sectors" ("realm_id");

CREATE INDEX ON "realm_sectors_jsonb" ("realm_sectors_jsonb_id");

CREATE INDEX ON "levies" ("realm_member_id");

CREATE INDEX ON "levies" ("realm_id");

CREATE INDEX ON "levies" ("encampment");

CREATE INDEX ON "levies" ("stationed");

CREATE INDEX ON "levies_actions" ("levy_id");

CREATE INDEX ON "levies_actions" ("expected_completion_at");

CREATE INDEX ON "levies_actions" ("completed");

CREATE INDEX ON "conquered_nations" ("owner_id");

CREATE INDEX ON "battle_outcome" ("levy_action_id");

CREATE INDEX ON "levy_surrenders" ("levy_id");

CREATE INDEX ON "levy_surrenders" ("receiving_realm_id");

ALTER TABLE "realm_members" ADD FOREIGN KEY ("user_id") REFERENCES "realms" ("owner_id");

ALTER TABLE "users" ADD FOREIGN KEY ("user_id") REFERENCES "realm_members" ("user_id") ON DELETE SET NULL;

ALTER TABLE "sectors" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "realms" ADD FOREIGN KEY ("realm_id") REFERENCES "realm_sectors_jsonb" ("realm_sectors_jsonb_id") ON DELETE CASCADE;

ALTER TABLE "levies" ADD FOREIGN KEY ("encampment") REFERENCES "sectors" ("cell_number");

ALTER TABLE "levies" ADD FOREIGN KEY ("realm_member_id") REFERENCES "realm_members" ("user_id");

ALTER TABLE "levies" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id") ON DELETE SET NULL;

ALTER TABLE "levies_actions" ADD FOREIGN KEY ("levy_id") REFERENCES "levies" ("levy_id") ON DELETE CASCADE;

ALTER TABLE "conquered_nations" ADD FOREIGN KEY ("owner_id") REFERENCES "realm_members" ("user_id");

ALTER TABLE "battle_outcome" ADD FOREIGN KEY ("levy_action_id") REFERENCES "levies_actions" ("levy_action_id") ON DELETE CASCADE;

ALTER TABLE "levies" ADD FOREIGN KEY ("levy_id") REFERENCES "levy_surrenders" ("levy_id") ON DELETE CASCADE;

ALTER TABLE "levy_surrenders" ADD FOREIGN KEY ("receiving_realm_id") REFERENCES "realms" ("realm_id");

-- 추가 key
ALTER TABLE "users" ADD CONSTRAINT "email_key" UNIQUE ("email");

ALTER TABLE "users" ADD CONSTRAINT "nickname_key" UNIQUE ("nickname");
