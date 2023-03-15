---
sidebar_label: 'Quick Start'
sidebar_position: 4
---

# Quick Start

:::caution
Make sure you have a public SSH key in `$HOME/.ssh/id_rsa.pub`
:::

## Create

1. Download an Ubuntu image :

```shell
wget 'https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img' \
    -O /tmp/ubuntu-22.04-server-cloudimg-amd64.img
```

2. Create the virtual machine :

```sh
cd freyja
freyja machine create -c examples/basic.yaml
freyja machine info
```

## List

Check the created virtual machines :

```sh
freyja machine list
```

Check the created machine information :

```sh
freyja machine info
```

:::tip Connexion
You may now connect to this machine using SSH connexion with the default user `freyja:master`.
:::

## Delete

Remove the created virtual machine :

```sh
freyja machine delete freyja-ubuntu
```
