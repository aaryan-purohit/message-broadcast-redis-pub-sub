output "redis_container_name" {
  value = docker_container.redis.name
}

output "redis_port" {
  value = docker_container.redis.ports[0].external
}