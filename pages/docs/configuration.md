---
sidebar_label: 'Configuration'
sidebar_position: 6
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
| image                   | Yes             | String       | None                               | Path of the image file to create your VM host. It will not be altered but copied in the virtual machine folder built by freyja.                                                     |
| os                      | Yes             | String       | None                               | Name of the OS to create your VM host. It **MUST** match the OS of the image file. Use the command `osinfo-query os` to get the list of the accepted OS names. Example : `centos8`. |
| hostname                | Yes             | String       | None                               | Name of the VM host. It **MUST NOT** contain underscores.                                                                                                                           |
| networks                | No              | List[Object] | `default`                          | List of networks. Each one of them will be used to create a new interface if it does not exist.                                                                                     |
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
| ignition.version        | Yes for section | String       | None                               | Version of the Ignition **specs**. Make sure this version fits with the `os` version for compatibility concerns.                                                                    |
| ignition.file           | No              | String       | None                               | Ignition file path for VM provisioning. If provided, this file will override the generated ignition provisioning file with its own content.                                         |
| update                  | No              | Boolean      | False                              | If `true`, update and upgrade the system after first boot.                                                                                                                          |
| reboot                  | No              | Boolean      | False                              | If `true`, reboot the machine after all boot tasks have completed.                                                                                                                  |

Example :

```yaml
# freyja version used to launch this configuration
version: "v0.1.0-beta"

hosts:
  # Mandatory hostname that must be unique
  # It defines the domain name of the machine in libvirt
  - hostname: "debian11"

    # Mandatory path to the image to start the virtual machine
    # It will not be altered but copied in the virtual machine folder built by Freyja.
    image: "$HOME/Images/debian-11-generic-amd64-20221020-1174.qcow2"

    # Mandatory OS type name
    # use 'osinfo-query os' to choose the proper os name to use
    os: "debian11"
    
    # Optional disk size for the sysroot partition of the guest's filesystem ('/')
    # Value in GB
    disk: 30
    
    # Optional memory size of the guest
    # Value in MB
    memory: 4096
    
    # Optional number of vcpus of the guest
    vcpus: 2

    # Optional List of unix accounts to create at boot
    # All of these accounts will be created alongside their home directory
    # The default is 'freyja:master' with no keys and no additional groups
    users:
        # Optional unix account name
        # will be created in a group of the same and with a home directory
      - username: "freyja"
        # Optional password value hashed in sha-256.
        # use mkpasswd -m sha-256 to define yours
        password: "$5$B.GSTERJDYuWJoV3$4koDNFgtRc137XVxw8jlcMyYo8lCbtyDT0bv.UA.ex6"
        # Optional paths of ssh keys to inject in 'authorized_keys' file on the final guest
        # useful to inject the ssh keys of your host during the guest boot phase
        keys: [ "$HOME/.ssh/id_ed25519.pub" ]
        # Optional additional groups for the current unix account
        groups: [ "sudo" ]

    # Optional list of subnetworks to create alongside the virtual machine in libvirt
    # Default network is 'default' and a dynamic MAC address will be provided by libvirt
    # By default, the defined networks will be attached to the libvirt bridge 'virbr0'
    networks:
        # Mandatory name of the network domain to create in libvirt
        # If this network does not exist yet, it will be created
        # If this network already exists, it will be attached to the current virtual machine 
      - name: "default"
        # Mandatory MAC address of the virtual machine inside the current subnetwork domain of libvirt
        # The address must be unique in this network
        address: "52:54:02:aa:aa:aa"
        
    # Optional list of files to write on the guest filesystem at boot
    write-files:
        # Mandatory source of the file on the host's filesystem
      - source: "/tmp/file.txt"
        # Mandatory destination of the file on the guest's filesystem
        destination: "/etc/file.txt"
        # Optional permissions of the final file on the guest's filesystem
        permissions: "0600"
        # Optional owner and group of the final file on the guest's filesystem
        owner: "root:root"
          
    # Optional configuration to update the guest system
    # Happens before additional packages installation 
    update: false
    
    # Optional list of additional packages to install at boot
    packages: ["vim"]
    
    # Optional list of commands to run at boot
    # Happens after additional files injection and additional packages installation
    # The default system's shell is used
    runcmd:
      - curl -fL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644
    
    # Optional configuration to reboot the gest at the very end of the boot phase
    reboot: false
    
    # Optional configuration for Ignition files
    # If configured with no ignition file, the OS will be provisioned with the current configurations
    #   defined above
    ignition:
      # Mandatory version of the Ignition's specification to use for provisioning the current guest
      version: "3.3.0"
      # Optional Ignition file to boot and provision the currest guest
      # If configured, the options defined in this Ignition file will override the current configurations
      #  defined above
      file: "$HOME/freyja-conf/ignition.json"
```
