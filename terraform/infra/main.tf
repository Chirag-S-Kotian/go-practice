# bucket to store website

resource "google_storage_bucket" "website" {
  name     = "chirag117-website117"
  location = "US"
}

# make new object public
resource "google_storage_object_access_control" "public_rule" {
  bucket = google_storage_bucket.website.name
  object = google_storage_bucket_object.index.name
  role   = "READER"
  entity = "allUsers"
}

# upload website files to bucket
resource "google_storage_bucket_object" "index" {
  name   = "index.html"
  bucket = google_storage_bucket.website.name
  source = "../website/index.html"
}

# reserve a static external IP address for the website
resource "google_compute_global_address" "website_ip" {
  name = "website-ip"
}

# get the managed DNS zone for the domain name
data "google_dns_managed_zone" "website_zone" {
  name = "chirag117-zone"
}

# add the ip to the DNS zone
resource "google_dns_record_set" "website_ip" {
  name         = "website.${data.google_dns_managed_zone.website_zone.dns_name}"
  managed_zone = data.google_dns_managed_zone.website_zone.name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_global_address.website_ip.address]
}
