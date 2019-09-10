# Tired of manually adding Proxmox VE nodes IPs in /etc/hosts?
This tool is what you need, just run `pve-hosts >> /etc/hosts` to get every PVE cluster node ip address on vmbr0(or any other, specify with -iface flag) interface in /etc/hosts

## Usage

#### Simple as 1-2-3
`pve-hosts [-iface <ifname>]`

This one (if it executed on cluster node) prints to stdout list of hosts with it ip addressess, if no iface flag specified, ip address of vmbr0 interface will be printed. If node have no ip on selected interface, it will not be included in list

Works great with PVE 5.3 and 5.4, i guess it will work on others 5.x installation. Currently i have no PVE 6 cluster installation to test compatibility, but it may work on v6 (but im not sure)

#### Warning

If node have multiple IP addressess on given interface, only first one will be included, for others specify interface with index (like vmbr0:1)
