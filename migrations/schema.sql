-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';
-- Create "households" table
CREATE TABLE "public"."households" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "owner_name" character varying(255) NOT NULL,
  "address" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "waste_pickups" table
CREATE TABLE "public"."waste_pickups" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "household_id" uuid NOT NULL,
  "type" character varying(20) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "pickup_date" timestamptz NULL,
  "safety_check" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_pickup_household" FOREIGN KEY ("household_id") REFERENCES "public"."households" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_pickup_household_id" to table: "waste_pickups"
CREATE INDEX "idx_pickup_household_id" ON "public"."waste_pickups" ("household_id");
-- Create index "idx_pickup_status" to table: "waste_pickups"
CREATE INDEX "idx_pickup_status" ON "public"."waste_pickups" ("status");
-- Create index "idx_pickup_type" to table: "waste_pickups"
CREATE INDEX "idx_pickup_type" ON "public"."waste_pickups" ("type");
-- Create "payments" table
CREATE TABLE "public"."payments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "household_id" uuid NOT NULL,
  "waste_id" uuid NOT NULL,
  "amount" numeric(10,2) NOT NULL,
  "payment_date" timestamptz NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "proof_file_url" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_household" FOREIGN KEY ("household_id") REFERENCES "public"."households" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_payment_pickup" FOREIGN KEY ("waste_id") REFERENCES "public"."waste_pickups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_payment_household_id" to table: "payments"
CREATE INDEX "idx_payment_household_id" ON "public"."payments" ("household_id");
-- Create index "idx_payment_status" to table: "payments"
CREATE INDEX "idx_payment_status" ON "public"."payments" ("status");
