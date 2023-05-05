---
title: 'Python requirements'
sidebar_label: 'Python & Poetry'
---

# Install Python 3 and Poetry

## On Ubuntu

Check python 3.9 :

```sh
python3 --version
```

If python 3.9 is not installed :

```sh
sudo apt install software-properties-common
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt install python3.9
```

Install Poetry :

```sh
# poetry is installed in $HOME/.local/bin
curl -sSL https://install.python-poetry.org | python3 -
export PATH="$HOME/.local/bin:$PATH"
```
