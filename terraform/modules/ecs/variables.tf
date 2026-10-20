variable "service_name" {
  type = string
}
variable "image" {
  type = string
}
variable "container_port" {
  type = number
}
variable "subnet_ids" {
  type = list(string)
}
variable "security_group_ids" {
  type = list(string)
}
variable "execution_role_arn" {
  type = string
}
variable "task_role_arn" {
  type = string
}
variable "log_group_name" {
  type = string
}
variable "ecs_count" {
  type    = number
  default = 1
}
variable "region" {
  type = string
}
variable "cpu" {
  type    = string
  default = "256"
}
variable "memory" {
  type    = string
  default = "512"
}
variable "target_group_arn" {
  type    = string
  default = ""
}
variable "min_capacity" {
  type    = number
  default = 2
}
variable "max_capacity" {
  type    = number
  default = 4
}
variable "cpu_target" {
  type    = number
  default = 70
}
