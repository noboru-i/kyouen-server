resource "aws_instance" "ecs-instance" {
    ami = "ami-b3afa2dd"
    instance_type = "t2.micro"
    key_name = "aws_key"
    security_groups = ["${aws_security_group.allow_all.id}"]
    monitoring = true
    subnet_id = "${aws_subnet.kyouen-subnet.id}"
    associate_public_ip_address = true
    user_data = "${file("aws_instance-user_data/userdata.sh")}"
    iam_instance_profile = "ec2_profile"
}

resource "aws_elb" "kyouen-elb" {
  name = "kyouen-elb"
  subnets = ["${aws_subnet.kyouen-subnet.id}"]
  security_groups = ["${aws_security_group.allow_all.id}"]

  listener {
    instance_port = 8080
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 2
    timeout = 3
    target = "HTTP:8080/"
    interval = 30
  }

  instances = ["${aws_instance.ecs-instance.id}"]
  cross_zone_load_balancing = true
  idle_timeout = 400
  connection_draining = true
  connection_draining_timeout = 400
}
