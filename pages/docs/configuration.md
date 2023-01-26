---
sidebar_label: 'Configuration'
sidebar_position: 5
---

# Configuration

:::info Examples
You can check configuration examples to start with :
* [A basic example](_static/basic.yaml)
* [A detailed example](_static/detailed.yaml)
* [An example for immutable OS](_static/immutable.yaml)
:::

The configuration definition consists in a configuration version and a list of hosts :

| **Configuration** | **Mandatory** | **Type** | **Default Value** | **Description**                           |
|-------------------|---------------|----------|-------------------|-------------------------------------------|
| version           | Yes           | String   | None              | Version of Freyja                         |
| hosts.[]          | Yes           | List     | None              | Configure each VM host you want to create |

For each host, you may configure :

| **Configuration**       | **Mandatory**   | **Type**     | **Default Value**                  | **Description**                                                                                                                                                                     |
|-------------------------|-----------------|--------------|------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| image                   | Yes             | String       | None                               | Path of the image file to create your VM host.                                                                                                                                      |
| os                      | Yes             | String       | None                               | Name of the OS to create your VM host. It **MUST** match the OS of the image file. Use the command `osinfo-query os` to get the list of the accepted OS names. Example : `centos8`. |
| hostname                | Yes             | String       | None                               | Name of the VM host. It **MUST NOT** contain underscores.                                                                                                                           |
| networks                | No              | List[Object] | `default`                          | List of networks. Each one of them will be used to create a new interface if it does not exist. You **MUST** configure at least one network.                                        |
| networks[].name         | Yes for section | String       | `default`                          | The name of the network.                                                                                                                                                            |
| networks[].address      | Yes for section | String       | Choosed by libvirt / external DHCP | MAC address of the current network for this host.                                                                                                                                   |
| users                   | No              | List[Object] | `freyja:master`                    | List of users. If none, a default `freyja:master` user is created without an authorized ssh key                                                                                     |
| users[].username        | No              | String       | `freyja`                           | Username to log into the created VM host.                                                                                                                                           |
| users[].password        | No              | String       | `master`                           | SHA2 hashed password to log into the created VM host. Use command `mkpasswd -m sha-512` to hash the password                                                                        |
| users[].keys[]          | No              | List[String] | None                               | Path to authorized ssh keys for this user.                                                                                                                                          |
| users[].groups[]        | No              | List[String] | None                               | Add the user into these configured groups.                                                                                                                                          |
| disk                    | No              | Integer      | `30`                               | Size of the VM disk in GigaBytes.                                                                                                                                                   |
| memory                  | No              | Integer      | `4096`                             | Size of the VM memory in MegaBytes.                                                                                                                                                 |
| vcpus                   | No              | Integer      | `2`                                | Number of VCPUS of the VM.                                                                                                                                                          |
| packages                | No              | List[String] | None                               | Additional packages to install during VM startup                                                                                                                                    |
| runcmd                  | No              | List[String] | None                               | Additional commands to run for the first boot of the machine                                                                                                                        |
| write-files             | No              | List[Object] | None                               | Additional files to write on machines at boot                                                                                                                                       |
| write-files.source      | Yes for section | String       | None                               | Source file with content to write on the machine                                                                                                                                    |
| write-files.destination | Yes for section | String       | None                               | Destination file on machine to write with the content of the source file                                                                                                            |
| write-files.permissions | No              | String       | `0600`                             | Permissions of the destination file on the machine                                                                                                                                  |
| write-files.owner       | No              | String       | `root:root`                        | Owner of the destination file on the machine. Format is `user:group`.                                                                                                               |
| ignition                | No              | Object       | None                               | If enabled, triggers the ignition provisioning file generation and disable the default cloud-init provisioning. Then starts the VMs with ignition provisioning.                     |
| ignition.version        | Yes for section | String       | None                               | Ignition version supported by the `os`. Make sure this version fits with the `os` version for compatibility concerns.                                                               |
| ignition.file           | No              | String       | None                               | Ignition file path for VM provisioning. If provided, this file will override the generated ignition provisioning file with its own content.                                         |
| update                  | No              | Boolean      | False                              | If `true`, update and upgrade the system after first boot.                                                                                                                          |
| reboot                  | No              | Boolean      | False                              | If `true`, reboot the machine after all boot tasks have completed.                                                                                                                  |
