# About this example

This Terraform configuration creates the defined number of OpenStack instances
and assigns floating IPs to each of them. Additionally, it will run a bootstrap
script that configures the OpenStack instances to run the mysql service.

Please read the openstack.tf file to make sure your options are setup correctly.
Pay special attention to the following variables:

variables.tf
* auth_url
* user_name

openstack.tf
* mysql_instance_count
* connection > user
* connection > private_key
* image_name
* instance_flavor

