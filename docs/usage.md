---
sidebar_label: 'Usage'
sidebar_position: 5
---

# Usage

:::info
To create virtual machines and networks, you need to :

* Download an OS cloud image (qcow2, img, ...)
* Create a YAML configuration
:::

The proper virtual host and networks described in the configuration file will be created using Libvirt and Qemu.

For further usage, run :

```sh
freyja --help
```

## Tips

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

Opens a console in a specific machine :

```sh
freyja machine console vm1
```

List mac addresses already in use :

```sh
freyja machine info | grep mac
```

Create a snapshot of a machine :
```sh
freyja machine snapshot vm1 snaphost_name
```

Restore a snapshot of a machine :
```sh
freyja machine restore vm1 snaphost_name
```

List the snapshots of a machine :
```sh
freyja machine list-snapshots vm1
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
