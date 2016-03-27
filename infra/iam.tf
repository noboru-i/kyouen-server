resource "aws_iam_instance_profile" "ec2_profile" {
    name = "ec2_profile"
    roles = ["${aws_iam_role.ec2_role.name}"]
}

resource "aws_iam_role" "ec2_role" {
    name = "ec2_role"
    assume_role_policy = "${file("aws_iam_role-assume_role_policy/ec2_role.json")}"
}

resource "aws_iam_role_policy" "AmazonEC2ContainerServiceforEC2Role" {
    name = "AmazonEC2ContainerServiceforEC2Role"
    role = "${aws_iam_role.ec2_role.id}"
    policy = "${file("aws_iam_role_policy-policy/AmazonEC2ContainerServiceforEC2Role.json")}"
}

resource "aws_iam_role" "ecs_role" {
    name = "ecs_role"
    assume_role_policy = "${file("aws_iam_role-assume_role_policy/ecs_role.json")}"
}

resource "aws_iam_role_policy" "AmazonEC2ContainerServiceRole" {
    name = "AmazonEC2ContainerServiceRole"
    role = "${aws_iam_role.ecs_role.id}"
    policy = "${file("aws_iam_role_policy-policy/AmazonEC2ContainerServiceRole.json")}"
}
