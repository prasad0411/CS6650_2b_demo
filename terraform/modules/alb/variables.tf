variable "service_name" { type = string }
variable "container_port" { type = number }
variable "vpc_id" { type = string }
variable "subnet_ids" { type = list(string) }
variable "alb_sg_id" { type = string }
