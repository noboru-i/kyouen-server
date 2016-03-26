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

resource "aws_instance" "ecs-instance" {
    ami = "ami-b3afa2dd"
    instance_type = "t2.micro"
    iam_instance_profile = "ec2_profile"
    user_data = "${file("user_data/userdata.sh")}"
}

resource "aws_iam_instance_profile" "ec2_profile" {
    name = "ec2_profile"
    roles = ["${aws_iam_role.ec2_role.name}"]
}

resource "aws_iam_role" "ec2_role" {
    name = "ec2_role"
    assume_role_policy = "${file("aws_iam_role_policies/ec2_role.json")}"
}

resource "aws_iam_role_policy" "AmazonEC2ContainerServiceforEC2Role" {
    name = "AmazonEC2ContainerServiceforEC2Role"
    role = "${aws_iam_role.ec2_role.id}"
    policy = "${file("aws_iam_group_policies/AmazonEC2ContainerServiceforEC2Role.json")}"
}

resource "aws_iam_role" "ecs_role" {
    name = "ecs_role"
    assume_role_policy = "${file("aws_iam_role_policies/ecs_role.json")}"
}

resource "aws_iam_role_policy" "AmazonEC2ContainerServiceRole" {
    name = "AmazonEC2ContainerServiceRole"
    role = "${aws_iam_role.ecs_role.id}"
    policy = "${file("aws_iam_group_policies/AmazonEC2ContainerServiceRole.json")}"
}

resource "aws_ecs_cluster" "kyouen-cluster" {
  name = "kyouen-cluster"
}

resource "aws_ecs_task_definition" "kyouen-registry" {
  family = "kyouen-registry"
  container_definitions = "${file("task-definitions/kyouen-registry.json")}"
}

resource "aws_elb" "kyouen-elb" {
  name = "kyouen-elb"
  availability_zones = ["ap-northeast-1a"]

  /*access_logs {
    bucket = "foo"
    bucket_prefix = "bar"
    interval = 60
  }*/

  listener {
    instance_port = 3000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 2
    timeout = 3
    target = "HTTP:8000/"
    interval = 30
  }

  instances = ["${aws_instance.ecs-instance.id}"]
  cross_zone_load_balancing = true
  idle_timeout = 400
  connection_draining = true
  connection_draining_timeout = 400
}

resource "aws_ecs_service" "kyouen-service" {
  name = "kyouen-service"
  cluster = "${aws_ecs_cluster.kyouen-cluster.id}"
  task_definition = "${aws_ecs_task_definition.kyouen-registry.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_role.arn}"
  depends_on = ["aws_iam_role_policy.AmazonEC2ContainerServiceRole"]

  load_balancer {
    elb_name = "${aws_elb.kyouen-elb.id}"
    container_name = "kyouen"
    container_port = 3000
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
