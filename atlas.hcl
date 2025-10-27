data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./cmd/loader",
  ]
}

data "composite_schema" "app" {
  # Load enum types first.
  schema "maxit" {
    url = "file://schema.sql"
  }
  # Then, load the GORM models.
  schema "maxit" {
    url = data.external_schema.gorm.url
  }
}

env "gorm" {
  src = [
    data.composite_schema.app.url,
  ]

  dev = "docker://postgres/17/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
