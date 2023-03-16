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
sudo usermod -aG libvirt-qemu $USER
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
