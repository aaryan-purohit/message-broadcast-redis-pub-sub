variable "redis_port" {
  type    = number
  default = 6379
}

variable "redis_docker_image" {
  type    = string
  default = "redis:8.2.4-alpine3.22"
}