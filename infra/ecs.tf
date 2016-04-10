resource "aws_ecs_cluster" "kyouen-cluster" {
  name = "kyouen-cluster"
}

resource "aws_ecs_task_definition" "kyouen-registry" {
  family = "kyouen-registry"
  container_definitions = "${file("aws_ecs_task_definition-container_definitions/kyouen.json")}"
}

resource "aws_ecs_service" "kyouen-service" {
  name = "kyouen-service"
  cluster = "${aws_ecs_cluster.kyouen-cluster.id}"
  task_definition = "${aws_ecs_task_definition.kyouen-registry.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_role.arn}"
  deployment_maximum_percent = 100
  deployment_minimum_healthy_percent = 0
  depends_on = ["aws_iam_role_policy.AmazonEC2ContainerServiceRole"]

  load_balancer {
    elb_name = "${aws_elb.kyouen-elb.id}"
    container_name = "kyouen"
    container_port = 3000
  }
}
