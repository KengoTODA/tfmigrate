tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "gcs" {
      bucket = "tfstate-test"
      name   = "tfmigrate/history.json"

      // mock gcs endpoint with fake-gcs-server
      endpoint = "http://fake-gcs-server:4443"
    }
  }
}
