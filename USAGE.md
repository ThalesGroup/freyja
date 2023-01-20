# USAGE

## SETUP

To create virtual machines, you need to :

* download an OS image
* create a YAML configuration

Check the configuration examples :

* [A basic configuration](examples/basic.yaml)
* [A more detailed configuration](examples/detailed.yaml)

For custom configuration, read the [machine configuration section](#machine-configuration).

The host is created with the proper virtual host and networks described in the configuration file.

For custom usage, use :

```sh
./freyja.sh --help
```

## CONFIGURATION

The configuration definition consists in a list of hosts :

| **Configuration** | **Mandatory** | **Type** | **Default Value** | **Description**                           |
|-------------------|---------------|----------|-------------------|-------------------------------------------|
| hosts.[]          | Yes           | List     | None              | Configure each VM host you want to create |

For each host, you may configure :

| **Configuration**  | **Mandatory**   | **Type**     | **Default Value**       | **Description**                                                                                                                                                                     |
|--------------------|-----------------|--------------|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| image              | Yes             | String       | None                    | Path of the image file to create your VM host.                                                                                                                                      |
| os                 | Yes             | String       | None                    | Name of the OS to create your VM host. It **MUST** match the OS of the image file. Use the command `osinfo-query os` to get the list of the accepted OS names. Example : `centos8`. |
| hostname           | Yes             | String       | None                    | Name of the VM host. It **MUST NOT** contain underscores.                                                                                                                           |
| networks           | No              | List[Object] | `default`               | List of networks. Each one of them will be used to create a new interface if it does not exist. You **MUST** configure at least one network.                                        |
| networks[].name    | Yes for section | String       | `default`               | The name of the network.                                                                                                                                                            |
| networks[].address | Yes for section | String       | Choosed by libvirt DHCP | MAC address of the current network for this host.                                                                                                                                   |
| users              | No              | List[Object] | `freyja:master`         | List of users. If none, a default `freyja:master` user is created without an authorized ssh key                                                                                     |
| users[].username   | No              | String       | `freyja`                | Username to log into the created VM host.                                                                                                                                           |
| users[].password   | No              | String       | `master`                | SHA2 hashed password to log into the created VM host. Use command `mkpasswd -m sha-512` to hash the password                                                                        |
| users[].keys[]     | No              | List[String] | None                    | Path to authorized ssh keys for this user.                                                                                                                                          |
| users[].groups[]   | No              | List[String] | None                    | Add the user into these configured groups.                                                                                                                                          |
| disk               | No              | Integer      | `30`                    | Size of the VM disk in GigaBytes.                                                                                                                                                   |
| memory             | No              | Integer      | `4096`                  | Size of the VM memory in MegaBytes.                                                                                                                                                 |
| vcpus              | No              | Integer      | `2`                     | Number of VCPUS of the VM.                                                                                                                                                          |
| packages           | No              | List[String] | None                    | Additional packages to install during VM startup                                                                                                                                    |
| ignition           | No              | Object       | None                    | If enabled, triggers the ignition provisioning file generation and disable the default cloud-init provisioning. Then starts the VMs with ignition provisioning.                     |
| ignition.version   | Yes for section | String       | None                    | Ignition version supported by the `os`. Make sure this version fits with the `os` version for compatibility concerns.                                                               |
| ignition.file      | No              | String       | None                    | Ignition file path for VM provisioning. If provided, this file will override the generated ignition provisioning file with its own content.                                         |


## IMMUTABLE OS IGNITION EXAMPLE

You may start and provision a virtual machine with an immutable OS like Flatcar (or Fedora CoreOS, ...).

If you run Freyja on Ubuntu, you will need to grant `read` permission to `libvirt` for ignition files directory.
Update `/etc/apparmor.d/abstractions/libvirt-qemu` with `read` permission for the Freyja workspace :

```sh
echo "  <home user>/freyja-workspace/** r," >> /etc/apparmor.d/abstractions/libvirt-qemu
sudo systemctl restart apparmor
```

Then use the [immutable_ignition.yaml configuration file example](./examples/flatcar.yaml) :

```sh
# stable flatcar image
wget https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2{,.sig} -P /tmp
bunzip2 /tmp/flatcar_production_qemu_image.img.bz2
# create the vm quickly
freyja machine create -c examples/flatcar.yaml
# check its state
freyja machine info
```

You may now connect to this machine using an SSH client using the info's IP address.  
The default user is `freyja:master`.

## CHEATSHEET

Debug the configuration :

```sh
freyja machine create -c ~/myconfig.yaml --dry-run -v
```

Install the machines using the configuration :

```sh
freyja machine create -c myconfig.yaml
```

List the existing machines :

```sh
freyja machine list
```

Describe the existing machines :

```sh
freyja machine info
# filter by name
freyja machine info vm1 vm2
```

Check machines' activity in real time :

```sh
# all
freyja machine usage --watch
# filter by name
freyja machine usage vm1 vm2 --watch
# static usage (one time display)
freyja machine usage vm1
```

List mac addresses already in use :

```sh
freyja machine info | grep mac
```

List networks:

```sh
freyja network list
```

Describe networks :

```sh
freyja network info 
# filter by name
freyja network info net1 net2
```