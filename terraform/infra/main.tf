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
  name = "luffy-terraform"
}

# add the ip to the DNS zone
resource "google_dns_record_set" "website_ip" {
  name         = "website.${data.google_dns_managed_zone.website_zone.dns_name}"
  managed_zone = data.google_dns_managed_zone.website_zone.name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_global_address.website_ip.address]
}

# add the bucket as a CDN backend
resource "google_compute_backend_bucket" "website_backend" {
  name        = "website-backend"
  bucket_name = google_storage_bucket.website.name
  description = "Backend for website"
  enable_cdn  = true
}

# create url map for the website
resource "google_compute_url_map" "website_map" {
  name            = "website-map"
  default_service = google_compute_backend_bucket.website_backend.self_link
  host_rule {
    hosts        = ["*"]
    path_matcher = "allpaths"
  }
  path_matcher {
    name            = "allpaths"
    default_service = google_compute_backend_bucket.website_backend.self_link
  }
}

# create a global HTTP(S) load balancer
resource "google_compute_target_http_proxy" "website_proxy" {
  name    = "website-proxy"
  url_map = google_compute_url_map.website_map.self_link
}

# create a global forwarding rule for the load balancer
resource "google_compute_global_forwarding_rule" "website_forwarding_rule" {
  name                  = "website-forwarding-rule"
  load_balancing_scheme = "EXTERNAL"
  ip_address            = google_compute_global_address.website_ip.address
  ip_protocol           = "TCP"
  port_range            = "80"
  target                = google_compute_target_http_proxy.website_proxy.self_link
}
