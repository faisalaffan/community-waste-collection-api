env "local" {
  src = "migrations/20250617000000_init.hcl"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/16/dev"
}
