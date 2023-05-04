---
sidebar_label: 'Quick Start'
sidebar_position: 4
---

# Quick Start

Create an ssh key `~/.ssh/id_rsa.pub` :

```sh
ssh-keygen -t rsa
```

Download an Ubuntu image :

```shell
wget 'https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img' \
    -O /tmp/ubuntu-22.04-server-cloudimg-amd64.img
```

Create the virtual machine :

```sh
cd freyja
freyja machine create -c examples/basic.yaml
```

Check the created virtual machines :

```sh
freyja machine list
freyja machine info
```

:::tip Connexion
You may now connect to this machine using SSH connexion with the default user `freyja:master`.
:::

Remove the created virtual machine :

```sh
freyja machine delete freyja-ubuntu
```

## More

::info Other examples
For more usecases, check the [examples of configuration in the Github Repository](https://github.com/ThalesGroup/freyja/tree/main/examples)
:::
