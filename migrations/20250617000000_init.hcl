schema "public" {}

table "households" {
  schema = schema.public
  column "id"         { type = uuid; default = sql("gen_random_uuid()") }
  column "owner_name" { type = varchar(255); null = false }
  column "address"    { type = text; null = false }
  column "created_at" { type = timestamptz; default = sql("now()") }
  column "updated_at" { type = timestamptz; default = sql("now()") }
  primary_key { columns = [column.id] }
}

table "waste_pickups" {
  schema = schema.public
  column "id"           { type = uuid; default = sql("gen_random_uuid()") }
  column "household_id" { type = uuid; null = false }
  column "type"         { type = varchar(20); null = false }
  column "status"       { type = varchar(20); default = "pending" }
  column "pickup_date"  { type = timestamptz; null = true }
  column "safety_check" { type = boolean; default = false }
  column "created_at"   { type = timestamptz; default = sql("now()") }
  column "updated_at"   { type = timestamptz; default = sql("now()") }
  primary_key { columns = [column.id] }
  foreign_key "fk_pickup_household" {
    columns     = [column.household_id]
    ref_columns = [table.households.column.id]
    on_delete   = CASCADE
  }
  index "idx_pickup_household_id" { columns = [column.household_id] }
  index "idx_pickup_status"       { columns = [column.status] }
  index "idx_pickup_type"         { columns = [column.type] }
}

table "payments" {
  schema = schema.public
  column "id"             { type = uuid; default = sql("gen_random_uuid()") }
  column "household_id"   { type = uuid; null = false }
  column "waste_id"       { type = uuid; null = false }
  column "amount"         { type = decimal(10,2); null = false }
  column "payment_date"   { type = timestamptz; null = true }
  column "status"         { type = varchar(20); default = "pending" }
  column "proof_file_url" { type = text; null = true }
  column "created_at"     { type = timestamptz; default = sql("now()") }
  column "updated_at"     { type = timestamptz; default = sql("now()") }
  primary_key { columns = [column.id] }
  foreign_key "fk_payment_household" {
    columns     = [column.household_id]
    ref_columns = [table.households.column.id]
    on_delete   = CASCADE
  }
  foreign_key "fk_payment_pickup" {
    columns     = [column.waste_id]
    ref_columns = [table.waste_pickups.column.id]
    on_delete   = CASCADE
  }
  index "idx_payment_household_id" { columns = [column.household_id] }
  index "idx_payment_status"       { columns = [column.status] }
}
