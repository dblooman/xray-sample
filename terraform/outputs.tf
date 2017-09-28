output "instance_ip" {
  value = "${aws_instance.xray_instance.public_ip}"
}
