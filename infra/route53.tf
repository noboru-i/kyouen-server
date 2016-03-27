resource "aws_route53_zone" "kyouen-net" {
  name = "xn--b6qvb.net"
}

resource "aws_route53_record" "www" {
  zone_id = "${aws_route53_zone.kyouen-net.zone_id}"
  name = "www.xn--b6qvb.net"
  type = "A"

  alias {
    name = "${aws_s3_bucket.main.website_domain}"
    zone_id = "${aws_s3_bucket.main.hosted_zone_id}"
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "web" {
  zone_id = "${aws_route53_zone.kyouen-net.zone_id}"
  name = "web.xn--b6qvb.net"
  type = "A"

  alias {
    name = "${aws_elb.kyouen-elb.dns_name}"
    zone_id = "${aws_elb.kyouen-elb.zone_id}"
    evaluate_target_health = true
  }
}
