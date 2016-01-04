provider "openstack" {
    auth_url    =   "${var.auth_url.rack2}"
    tenant_name =   "${var.tenant_name}"
    user_name   =   "${var.user_name.syseng.simar}"
}

# The amount of instances you would like to spin
variable "mysql_instance_count" {
    default = 1
}

resource "openstack_compute_instance_v2" "mysql" {
    name            =   "${concat("mysql", count.index)}"
    image_name      =   "${var.image_name.ubuntu-trusty}"
    flavor_name     =   "${var.instance_flavor.medium}"
    region          =   "${var.region}"
    key_pair        =   "${var.user_name.syseng.simar}"
    security_groups =   [
        "default",
        "mysql",
        "web"
    ]
    floating_ip     =   "${element(openstack_compute_floatingip_v2.fip.*.address, count.index)}"
    count           =   "${var.mysql_instance_count}"

    # Your ssh key here
    connection {
        user        =   "ubuntu"
        private_key =   "${file(concat("/home/",var.user_name.syseng.simar,"/.ssh/id_rsa"))}"
    }

    # Setup the instances with the bootstrap script
    provisioner "file" {
        source      =   "./bootstrap.sh"
        destination =   "/tmp/bootstrap.sh"
    }

    # Run the instances with the bootstrap script
    provisioner "remote-exec" {
        inline = [
            "/bin/bash /tmp/bootstrap.sh"
        ]
    }
}

# A resource to create floating IPs
resource "openstack_compute_floatingip_v2" "fip" {
    region  =   "${var.region}"
    pool    =   "${var.external_network_pool.syseng}"
    count   =   "${var.mysql_instance_count}"
}



