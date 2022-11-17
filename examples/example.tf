terraform {
  required_version = ">= 1.1.5"
  required_providers {
    buildonaws = {
      source  = "aws.amazon.com/terraform/buildonaws"
    }
  }
}

provider "buildonaws" {
}

data "buildonaws_character" "deadpool" {
  identity = "Wade Wilson"
}

resource "buildonaws_character" "daredevil" {
  fullname = "Daredevil"
  identity = "Matt Murdock"
  knownas = "The man without fear"
  type = "super-hero"
}

output "daredevil_secret_identity" {
  value = "The secret identity of ${buildonaws_character.daredevil.fullname} is '${buildonaws_character.daredevil.identity}'"
}

output "deadpool_is_knownas" {
  value = "${data.buildonaws_character.deadpool.fullname} is also known as '${data.buildonaws_character.deadpool.knownas}'"
}
