CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "nickname" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realms" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "owner_id" bigint NOT NULL,
  "capital_number" int UNIQUE NOT NULL,
  "political_entity" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sectors" (
  "cell_number" int PRIMARY KEY,
  "province_number" int NOT NULL,
  "realm_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("email");

CREATE INDEX ON "realms" ("owner_id");

ALTER TABLE "realms" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id");

ALTER TABLE "sectors" ADD FOREIGN KEY ("realm_id") REFERENCES "realms" ("id");

ALTER TABLE "users" ADD CONSTRAINT "email_key" UNIQUE ("email");

ALTER TABLE "users" ADD CONSTRAINT "nickname_key" UNIQUE ("nickname");
