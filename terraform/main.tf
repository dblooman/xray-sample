provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "${var.region}"
}

data "aws_ami" "ubuntu_ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"]
}

resource "aws_instance" "xray_instance" {
  ami                  = "${data.aws_ami.ubuntu_ami.id}"
  instance_type        = "t2.micro"
  subnet_id            = "${var.subnet_id}"
  iam_instance_profile = "${aws_iam_instance_profile.xray.id}"
  user_data            = "${file("./resources/xray.sh")}"

  vpc_security_group_ids = [
    "${aws_security_group.xray.id}",
  ]

  key_name = "${var.key_name}"

  tags {
    Name = "X-Ray"
  }
}

resource "aws_security_group" "xray" {
  vpc_id = "${var.vpc_id}"
  name   = "Xray Demo"

  tags {
    Name = "Xray"
  }
}

resource "aws_security_group_rule" "http_in" {
  type      = "ingress"
  from_port = 8081
  to_port   = 8083
  protocol  = "tcp"

  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.xray.id}"
}

resource "aws_security_group_rule" "ssh_in" {
  type      = "ingress"
  from_port = 22
  to_port   = 22
  protocol  = "tcp"

  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.xray.id}"
}

resource "aws_security_group_rule" "http_web_out" {
  type      = "egress"
  from_port = 0
  to_port   = 0
  protocol  = "-1"

  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.xray.id}"
}

resource "aws_iam_instance_profile" "xray" {
  name = "xray-demo"
  role = "${aws_iam_role.xray.name}"
}

resource "aws_iam_role" "xray" {
  name = "xray_demo"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "xray_managed_policy" {
  name       = "xray_managed_policy_attachment"
  roles      = ["${aws_iam_role.xray.name}"]
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
}
