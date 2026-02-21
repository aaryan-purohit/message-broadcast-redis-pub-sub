terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.6.2"
    }
  }
}

provider "docker" {}

resource "docker_image" "redis" {
  name         = var.redis_docker_image
  keep_locally = false
}

resource "docker_container" "redis" {
  name  = "redis"
  image = docker_image.redis.image_id

  ports {
    internal = 6379
    external = var.redis_port
  }

  command = [
    "redis-server",
    "--appendonly",
    "yes"
  ]

  restart = "unless-stopped"
}