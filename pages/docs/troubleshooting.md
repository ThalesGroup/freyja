---
sidebar_label: 'Troubleshooting'
sidebar_position: 6
---

# Troubleshooting

## libvirt-qemu permission denied on home user

Creating a VM, you encounter the following error :

```sh
WARNING  /home/user/freyja-workspace/build/almalinux/almalinux_cloud_init.iso may not be accessible by th
e hypervisor. You will need to grant the 'libvirt-qemu' user search permissions for the following directo
ries: ['/home/user']                                
ERROR    Cannot access storage file '/home/user/freyja-workspace/build/image/image_snapshot' (as 
uid:64055, gid:109): Permission denied                                                                   
Domain installation does not appear to have been successful.
```

To solve this issue, set ACL of your host :

```sh
setacl u:libvirt-qemu:r $HOME
getacl -e $HOME

# file: home/user
# owner: user
# group: user
user::rwx
user:libvirt-qemu:r             #effective:r
group::r-x                      #effective:r-x
mask::r-x
other::---
```

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