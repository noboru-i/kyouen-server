resource "aws_s3_bucket" "main" {
    bucket = "www.xn--b6qvb.net"
    acl = "public-read"
    policy = "${file("aws_s3_bucket-policy/main.json")}"

    website {
        index_document = "index.html"
        error_document = "error.html"
    }
}
