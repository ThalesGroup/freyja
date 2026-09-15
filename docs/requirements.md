---
sidebar_label: 'Requirements'
sidebar_position: 2
---

# Requirements

## Runtime

Install Python >= 3.9

## Virtualization

Check your system requirements :

```sh
sudo apt install cpu-checker -y && kvm-ok
```

Install Qemu-KVM and Libvirt :

```sh
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils virtinst cloud-utils
# check installation
sudo systemctl is-active libvirtd
```

Allow your user to use Libvirt and Qemu-KVM:

```sh
sudo usermod -aG libvirt,kvm $USER
```

**Logout and login again !**

Check your groups permissions :

```sh
# libvirt & kvm MUST appear in your user's groups
groups
```

Verify that `virbr0` was created in bridges :

```sh
brctl show
```

## Rootless setup

Freyja is rootless : the machines run unprivileged in the user's libvirt **session**
daemon (`qemu:///session`), so their disk images stay in the user's home directory,
with the user's own permissions.

The virtual networks cannot run in a session daemon (creating a bridge, a DHCP
server and the firewall rules requires root), so Freyja manages the networks in the
privileged **system** daemon (`qemu:///system`) and shares them between the users.
This also rules out IP conflicts between users : each network has a single DHCP
server delivering unique leases, and libvirt refuses to start a network whose
subnet collides with an existing one.

The machines plug their interfaces into the networks' host bridges through
`qemu-bridge-helper`. To allow this, a host admin must, **once per host** :

```sh
# 1. grant the users access to the system daemon for the network management
sudo usermod -aG libvirt <user>

# 2. make qemu-bridge-helper setuid root, restricted to the 'libvirt' group
#    members (dpkg-statoverride persists the change across package upgrades)
sudo dpkg-statoverride --update --add root libvirt 4750 /usr/lib/qemu/qemu-bridge-helper

# 3. authorize the bridges freyja uses : 'virbr0' for the stock 'default'
#    network, and the predictable 'freyja0' to 'freyja9' range for the
#    networks freyja creates on the fly
sudo mkdir -p /etc/qemu { echo 'allow virbr0'; for i in $(seq 0 9); do echo "allow freyja$i"; done; } | sudo tee /etc/qemu/bridge.conf
```

:::warning

Do **not** use `allow all` in `/etc/qemu/bridge.conf` : it would let the
`libvirt` group members plug into **any** host bridge, including bridges
attached to physical interfaces.

:::

After this setup, freyja users never need root : `freyja machine create` builds
the missing networks on the fly through their `libvirt` group membership, and
starts the machines as their own user.
