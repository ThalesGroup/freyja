---
title: 'Virtualization requirements'
sidebar_label: 'KVM, Qemu & Libvirt'
---

# INSTALL QEMU-KVM & LIBVIRT

## Ubuntu

Check your system requirements :

```sh
sudo apt install cpu-checker -y && kvm-ok
```

Install qemu-kvm and libvirt :

```sh
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils virtinst cloud-utils
# check installation
sudo systemctl is-active libvirtd
```

Allow your user to use libvirt and kvm:

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
