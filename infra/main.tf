// env
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY

provider "aws" {
  region = "ap-northeast-1"
}

resource "aws_route53_zone" "kyouen-net" {
  name = "xn--b6qvb.net"
}

resource "aws_vpc" "kyouen-vpc" {
  cidr_block = "10.1.0.0/16"
  instance_tenancy = "default"
  enable_dns_support = "true"
  enable_dns_hostnames = "false"
  tags {
    Name = "kyouen-vpc"
  }
}

resource "aws_s3_bucket" "main" {
    bucket = "www.xn--b6qvb.net"
    acl = "public-read"
    policy = <<EOF
{"Statement":[{"Action":"s3:GetObject","Effect":"Allow","Principal":"*","Resource":"arn:aws:s3:::www.xn--b6qvb.net/*","Sid":"PublicReadForGetBucketObjects"}],"Version":"2012-10-17"}
EOF

    website {
        index_document = "index.html"
        error_document = "error.html"
    }
}

resource "aws_route53_record" "www" {
  zone_id = "${aws_route53_zone.kyouen-net.zone_id}"
  name = "www.xn--b6qvb.net"
  type = "A"

  alias {
    name = "s3-website-ap-northeast-1.amazonaws.com"
    zone_id = "${aws_s3_bucket.main.hosted_zone_id}"
    evaluate_target_health = true
  }
}

/*output "name_servers.0" {
  value = "${aws_route53_zone.kyouen-net.name_servers.0}"
}
output "name_servers.1" {
  value = "${aws_route53_zone.kyouen-net.name_servers.1}"
}
output "name_servers.2" {
  value = "${aws_route53_zone.kyouen-net.name_servers.2}"
}
output "name_servers.3" {
  value = "${aws_route53_zone.kyouen-net.name_servers.3}"
}*/
