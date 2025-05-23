# bucket to store website

resource "google_storage_bucket" "website" {
  name     = "chirag117-website117"
  location = "asia-south1"
}

# make new object public
resource "google_storage_object_access_control" "public_rule" {
  bucket = google_storage_bucket.website.name
  object = google_storage_bucket_object.index
  role   = "READER"
  entity = "allUsers"
}

# upload website files to bucket
resource "google_storage_bucket_object" "index" {
  name   = "index.html"
  bucket = google_storage_bucket.website.name
  source = "../website/index.html"
}
