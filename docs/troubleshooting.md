---
sidebar_label: 'Troubleshooting'
sidebar_position: 7
---

# Troubleshooting

## Cannot open Qemu system connexion

Running any freyja command, you encounter the following error :

```sh
Could not open Qemu system connexion for the networks (the user must be a member of the 'libvirt' group)
```

Freyja manages the virtual networks in the privileged libvirt system daemon,
which only the members of the `libvirt` group can reach. Ask your host admin to
run :

```sh
sudo usermod -aG libvirt <user>
```

Then logout and login again.

## Machine fails to start : bridge helper errors

`freyja machine create` stops before creating anything with an error such as :

```sh
ERROR: the host is not set up to let the machines plug into the host bridges {"reason":"invalid qemu-bridge-helper ACL file: open /etc/qemu/bridge.conf: no such file or directory", ...}
ERROR: the host is not set up to let the machines plug into the host bridges {"reason":"the qemu-bridge-helper ACL file '/etc/qemu/bridge.conf' does not allow the host bridges [freyja0] ...", ...}
```

or freyja warns :

```sh
WARN: qemu-bridge-helper is not setuid: the machines may not plug into the host bridges
```

or Qemu fails to start a machine with an error such as :

```sh
failed to communicate with bridge helper: access denied by acl file
failed to communicate with bridge helper: stderr=failed to parse default acl file `/etc/qemu/bridge.conf'
```

The machines run unprivileged and plug their interfaces into the networks' host
bridges through `qemu-bridge-helper`, which requires a one-time setup by the
host admin. See the requirements page, section **Rootless setup** :

```sh
sudo dpkg-statoverride --update --add root libvirt 4750 /usr/lib/qemu/qemu-bridge-helper
sudo mkdir -p /etc/qemu
echo 'allow <bridge>' | sudo tee -a /etc/qemu/bridge.conf   # for each freyja bridge
```

A bridge is allowed when an `allow` rule matches it and no `deny` rule does,
whatever their order in the file (`include` rules are followed).

With `--dry-run`, freyja only warns about the ACL file instead of stopping.

## Network not found

Creating a machine, you encounter the following error :

```sh
ERROR: cannot start the machine {"machine":"...","reason":"Network not found: no network with matching name 'default'"}
```

The machine was defined with an old freyja version that referenced the networks
by name in the session daemon. Delete the machine and create it again : the
interfaces now plug into the host bridges of the system daemon's networks.

Note that `virsh` targets the system daemon by default : to inspect the machines
freyja created in your session daemon, use `virsh -c qemu:///session` instead.

## Qemu unexpectedly closed the monitor

These kind of issues are often due to Apparmor or Selinux.

### Apparmor

Start by studying the error :

```sh
sudo journalctl -xe
```

You might encounter several kind of issues related to apparmor.

**CASE 1**

Output of `journalctl` error :

```sh
apparmor="DENIED" operation="file_mmap" profile="virt-aa-helper" name="/opt/openssl/lib/libcrypto.so
```

To solve this, you should grant `mmap` permission to `virt-aa-helper` for `libcrypto.so`.  
Update `/etc/apparmor.d/usr.lib.libvirt.virt-aa-helper` with `mmap` permission for `libcrypto` :

```sh
echo "  /opt/openssl/lib/libcrypto.so.1.1 m," >> /etc/apparmor.d/usr.lib.libvirt.virt-aa-helper
sudo systemctl restart apparmor
```

**CASE 2**

Output of `journalctl` error :

```sh
apparmor="DENIED" operation="open" profile="libvirt-2938f233-2589-4f85-9aa8-2f1cabd92dbf" name="~/freyja-workspace/build/myvm/provisioning.ign" pid=11837 comm="qemu-system-x86" requested_mask="r" denied_mask="r" fsuid=64055 ouid=1000
```

To solve this, you should grant `read` permission for `libvirt` for ignition files directory.
Update `/etc/apparmor.d/abstractions/libvirt-qemu` with `read` permission for the Freyja workspace :

```sh
echo "  /home/<user>/freyja-workspace/** r," >> /etc/apparmor.d/abstractions/libvirt-qemu
sudo systemctl restart apparmor
```

## IP is 'unknown'

The first reason might be that the virtual machine is shut off. Start it to verify the IP.

The second reason might be that the bridge is taking time to mount the vm IP on the host interface.  
Wait a few seconds and check again.

The third reason might be caused by an external DHCP resolution.  
Freyja is only capable of deducing IP addresses resolved on interfaces that are related to DHCP resolution from
libvirt on your local host. If the interface you are using for the current machine is not related to libvirt, the IP
resolution is made by an external server and Freyja is not able to give this information to you.

## Interface is 'unknown'

The first reason might be that the virtual machine is shut off. Start it to verify the interface.

## OS versions are missing in `osinfo-query os`

You need to update your osinfo database :

1. Install the package `osinfo-db-tools` on your system
2. Download the last version of the `osinfo-db-<version>.tar.xz` database. You may check the last version by visiting the [libosinfo index site \[?\]](https://releases.pagure.org/libosinfo/) :

```sh
wget -O "/tmp/osinfo-db.tar.xz" "https://releases.pagure.org/libosinfo/osinfo-db-20220214.tar.xz"
sudo osinfo-db-import --local /tmp/osinfo-db.tar.xz
osinfo-query os  # the os list should be updated 
```

## PolicyKit

libvirt has a policy in `/usr/share/polkit-1/rules.d/60-libvirt.rules` that allows the users taking part to the group
`libvirt` to manage virtual machines.

## Run tests

```sh
poetry run pytest --cov=freyja freyja/tests/
```

## Run build

Install Poetry.

Then run :

```sh
./build.sh
```

## Debug a cloud-init file

```sh
sudo apt install cloud-init
cloud-init schema -c ./user-data --annotate
```

## Manual debugging of machine files

```sh
freyja machine create -c myconfig.yaml --dry-run
cd $HOME/.freyja/machines/mymachine
# to regenerate cloud init config and iso image
genisoimage -output mymachine-cloud-init.iso -volid cidata -rock user-data meta-data
# define the machine in libvirt with the generated xml description
virsh define mymachine-libvirt-conf.xml
# start the domain in libvirt
virsh start mymachine
virsh list --all
```

## Restore default libvirt network

Dump into a file `default.xml` :

```sh
echo "<network>
    <name>default</name>
    <bridge name='virbr0'/>
    <forward/>
    <ip address='192.168.122.1' netmask='255.255.255.0'>
      <dhcp>
        <range start='192.168.122.2' end='192.168.122.254'/>
      </dhcp>
    </ip>
  </network>" >> /tmp/default.xml
```

Define it and push in libvirt :

```sh
virsh net-define /tmp/default.xml
virsh net-start /tmp/default
virsh net-autostart /tmp/default 
```