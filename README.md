# Tired of manually adding Proxmox VE nodes IPs in /etc/hosts?
This tool is what you need, just run `pve-hosts >> /etc/hosts` to get every PVE cluster node ip address on vmbr0(or any other, specify with -iface flag) interface in /etc/hosts

## Build status
Master branch: [![Build Status](https://dev.azure.com/niasar/pve-hosts/_apis/build/status/niasar.pve-hosts?branchName=master)](https://dev.azure.com/niasar/pve-hosts/_build/latest?definitionId=2&branchName=master)

Latest release binary: [![Get latest binary](https://img.shields.io/badge/Version-1.2.1-green.svg)](https://github.com/niasar/pve-hosts/releases/latest/download/pve-hosts)

## Usage

#### Simple as 1-2-3
`pve-hosts [-iface <ifname>]`

This one (if it executed on cluster node) prints to stdout list of hosts with it ip addresses, if no iface flag specified, ip address of vmbr0 interface will be printed. If node have no ip on given interface, it will not be included in list

Works great with PVE 5.3 and 5.4, i guess it will work on other 5.x installations. Currently i have no PVE v6 cluster to test compatibility, so it may work on v6, but im not sure

Also you can generate list in ansible hosts file format by specifying `-format ansible` flag

#### Warning

If node have multiple IP addressess on given interface, only first one will be included, for others specify interface with index (ex. vmbr0:1)

You need all your cluster nodes to be online, to generate host list 

Also if you have issues with your /etc/ssh/ssh_known_hosts file, execution will fail

### Execution time

Execution time depends on number of nodes in cluster. This tool is using pvesh to access proxmox API, every request takes approx. 2s to finish on idling node, number of requests needed to build the list is ![](https://latex.codecogs.com/gif.latex?n_{req}=n_{nodes}&plus;1). For now requests executing one by one, but maybe later i will make info gathering asynchronous to reduce execution time, or you can do it yourself via PR or forking :)

Have a nice day!