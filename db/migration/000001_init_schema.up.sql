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
  "capital_number" int UNIQUE NOT NULL,
  "political_entity" varchar NOT NULL,
  "color" varchar NOT NULL,
  "population_growth_rate" float NOT NULL,
  "state_coffers" int NOT NULL,
  "census_at" timestamptz NOT NULL,
  "tax_collection_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realm_members" (
  "user_id" bigint UNIQUE NOT NULL,
  "realm_id" bigint NOT NULL,
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
  "name" varchar NOT NULL,
  "morale" smallint NOT NULL,
  "encampment" int NOT NULL,
  "swordmen" int NOT NULL,
  "shield_bearers" int NOT NULL,
  "archers" int NOT NULL,
  "lancers" int NOT NULL,
  "supply_troop" int NOT NULL,
  "movement_speed" float NOT NULL,
  "offensive_strength" int NOT NULL,
  "defensive_strength" int NOT NULL,
  "realm_member_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("email");

CREATE INDEX ON "realms" ("owner_id");

CREATE INDEX ON "realm_members" ("realm_id");

CREATE INDEX ON "realm_members" ("user_id");

CREATE INDEX ON "sectors" ("realm_id");

CREATE INDEX ON "realm_sectors_jsonb" ("realm_sectors_jsonb_id");

CREATE INDEX ON "levies" ("realm_member_id");

ALTER TABLE "realms" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("user_id");

ALTER TABLE "realm_members" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");

ALTER TABLE "realm_members" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "sectors" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "realm_sectors_jsonb" ADD FOREIGN KEY ("realm_sectors_jsonb_id") REFERENCES "realms" ("realm_id");

ALTER TABLE "levies" ADD FOREIGN KEY ("realm_member_id") REFERENCES "realm_members" ("user_id");

-- 추가 key
ALTER TABLE "users" ADD CONSTRAINT "email_key" UNIQUE ("email");

ALTER TABLE "users" ADD CONSTRAINT "nickname_key" UNIQUE ("nickname");
