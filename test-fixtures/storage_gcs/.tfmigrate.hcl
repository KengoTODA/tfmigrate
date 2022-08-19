tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "gcs" {
      bucket = "tfstate-test"
      key    = "tfmigrate/history.json"

      // mock gcs endpoint with fake-gcs-server
      endpoint    = "http://localhost:4443"
      credentials = "dummy"
      name        = "tfmigrate/history.json"
    }
  }
}
