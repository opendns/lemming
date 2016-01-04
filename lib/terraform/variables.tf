# This is the authorization URL for your OpenStack server
variable "auth_url" {
    default = {
        "rack2" = "http://10.201.10.20:5000/v2.0"
    }
}

variable "tenant_name" {
    default = "syseng"
}

variable "user_name" {
    default = {
        "syseng.simar" = "simar"
    }
}

variable "region" {
    default = "regionOne"
}

variable "image_name" {
    default = {
        "ubuntu-trusty" = "ubuntu-1404-trusty-puppet"
    }
}

variable "instance_flavor" {
    default = {
        "tiny"      = "m1.tiny"
        "small"     = "m1.small"
        "medium"    = "m1.medium"
        "large"     = "m1.large"
        "xlarge"    = "m1.xlarge"
        "2xlarge"   = "m1.2xlarge"
    }
}

variable "external_network_pool" {
    default = {
        "generic"   = "ext-net"
        "syseng"    = "ext-syseng-net"
    }
}
