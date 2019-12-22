# connmgr
A ssh session manage tool under terminal

## Pros
* fuzzy search for creating and deleting record
* Manage multiple ssh session easily

## Help
```
Usage:
  connMgr [command]

Available Commands:
  a           Add a connection entry
  back        Backup connection config
  c           Connect to a host
  cmd         Send remote command to a host
  cp          Copy file to a host
  e           Enable root ssh access to a host
  help        Help about any command
  icp         Copy file from a host
  l           list all connections
  pub         dump public key
  r           remove a connections
  vim         Use vim to edit config

Flags:
  -h, --help   help for connMgr
```

## Dependency
* go 1.12
* fzf
  * Do not have library version
  * Install fzf system-wise (follow steps in https://github.com/junegunn/fzf)

## Install
```
git clone https://github.com/mrsandwitch/connmgr
cd connmgr; go build
```
