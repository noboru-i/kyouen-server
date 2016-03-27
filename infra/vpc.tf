resource "aws_vpc" "kyouen-vpc" {
  cidr_block = "10.0.0.0/16"
  instance_tenancy = "default"
  enable_dns_support = "true"
  enable_dns_hostnames = "false"
  tags {
    Name = "kyouen-vpc"
  }
}

resource "aws_internet_gateway" "kyouen-gw" {
  vpc_id = "${aws_vpc.kyouen-vpc.id}"

  tags {
    Name = "kyouen-gw"
  }
}

resource "aws_subnet" "kyouen-subnet" {
  vpc_id = "${aws_vpc.kyouen-vpc.id}"
  cidr_block = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags {
    Name = "kyouen-subnet"
  }
}

resource "aws_route_table" "kyouen-vpc-public-rt" {
  vpc_id = "${aws_vpc.kyouen-vpc.id}"
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.kyouen-gw.id}"
  }
  tags {
    Name = "kyouen-vpc-public-rt"
  }
}

resource "aws_route_table_association" "kyouen-vpc-rta" {
  subnet_id = "${aws_subnet.kyouen-subnet.id}"
  route_table_id = "${aws_route_table.kyouen-vpc-public-rt.id}"
}

resource "aws_security_group" "allow_all" {
  vpc_id = "${aws_vpc.kyouen-vpc.id}"
  name = "allow-all"
  description = "Allow all inbound traffic"
  ingress {
    from_port = 0
    to_port = 65535
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags {
    Name = "allow-all"
  }
}
