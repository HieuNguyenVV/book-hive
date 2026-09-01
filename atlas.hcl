variable "database_url" {
  type        = string
  # Host port mapped in docker-compose.yml (5439 -> 5432).
  default     = "postgres://postgres:postgres@localhost:5439/book_hive?sslmode=disable"
  description = "PostgreSQL connection URL (override with --var database_url=...)"
}

locals {
  migration_dir = "file://migrations"
  schema_src    = "file://schema/schema.sql"
  # Ephemeral Postgres used by Atlas to compute and validate diffs.
  dev_url       = "docker://postgres/16/dev?search_path=public"
}

env "local" {
  src = local.schema_src
  url = var.database_url
  dev = local.dev_url

  migration {
    dir = local.migration_dir
  }
}

# Used by migrator image / CI (set DATABASE_URL in the environment).
env "remote" {
  src = local.schema_src
  url = getenv("DATABASE_URL")

  migration {
    dir = local.migration_dir
  }
}
