CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "nickname" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "realms" (
  "id" bigserial PRIMARY KEY,
  "owner" bigint NOT NULL,
  "capital" bigint UNIQUE,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sectors" (
  "cell_id" int PRIMARY KEY,
  "province" int NOT NULL,
  "realm" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("email");

ALTER TABLE "realms" ADD FOREIGN KEY ("owner") REFERENCES "users" ("id");

ALTER TABLE "realms" ADD FOREIGN KEY ("capital") REFERENCES "sectors" ("cell_id");

ALTER TABLE "sectors" ADD FOREIGN KEY ("realm") REFERENCES "realms" ("id");

ALTER TABLE "users" ADD CONSTRAINT "email_key" UNIQUE ("email");

ALTER TABLE "users" ADD CONSTRAINT "nickname_key" UNIQUE ("nickname");