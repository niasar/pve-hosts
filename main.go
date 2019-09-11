package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	iface := flag.String("iface", "vmbr0", "network interface")
	flag.Parse()
	detectPveCluster()
	_, _ = fmt.Fprintln(os.Stderr, "Gathering info about cluster nodes...")
	list := getNodeNamelist()
	ipAddr := getIPaddressess(list, *iface)
	printResult(ipAddr, list)
}

// Accessing Proxmox API to get list of nodes

func getNodeNamelist() []string {
	var list []string
	var v interface{}
	jsonRaw, err := exec.Command("pvesh", "get", "nodes", "-output-format", "json").CombinedOutput()
	check(err)
	json.Unmarshal(jsonRaw, &v)
	v1 := v.([]interface{})
	for i := range v1 {
		data := v1[i].(map[string]interface{})
		for k, v := range data {
			if k == "node" {
				list = append(list, v.(string))
			}
		}
	}
	return list
}

//Getting /nodes/{$nodename}/network json for every node and iterating through it

func getIPaddressess(nodeList []string, iface string) map[string]string {
	var ipMap map[string]string
	var v interface{}
	var ifmap map[string]string
	ipMap = make(map[string]string)
	for i := range nodeList {
		ifmap = make(map[string]string)
		nodeName := nodeList[i]
		_, _ = fmt.Fprintln(os.Stderr, strings.Join([]string{"[", strconv.Itoa(i + 1), "/", strconv.Itoa(len(nodeList)), "]"}, ""), "Gathering info about", nodeList[i], "\b...")
		apiPath := strings.Join([]string{"/nodes/", nodeName, "/network"}, "")
		jsonRaw, err := exec.Command("pvesh", "get", apiPath, "-output-format", "json").CombinedOutput() // Accessing proxmox API about interfaces on node
		check(err)
		json.Unmarshal(jsonRaw, &v)
		v1 := v.([]interface{}) // Getting array of JSONs with network interfaces information here
		for i := range v1 {     // And iterating through it
			data := v1[i].(map[string]interface{}) // Getting interfaces name and ip
			var addr string
			var ifname string
			for k, v := range data {
				if k == "address" {
					addr = v.(string)
				}
				if k == "iface" {
					ifname = v.(string)
				}
			}
			ifmap[ifname] = addr // Writing data to map
		}
		if ipAddr, ok := ifmap[iface]; ok { // If interface provided with flag exist, write it to final map
			ipMap[nodeList[i]] = ipAddr
		}
	}
	return ipMap
}

func printResult(ipMap map[string]string, nodeList []string) {
	fmt.Fprintln(os.Stderr, "Printing result...")
	var longest int
	longest = 0
	for i := range nodeList {
		if len(nodeList[i]) > longest {
			longest = len(nodeList[i])
		}
	}
	maxDivider, _ := math.Modf(float64(longest) / 8)
	for i := range ipMap {
		var ipAddr string
		tabulator := "\t"
		divider, _ := math.Modf(float64(len(i)) / 8)
		tabs := maxDivider - divider
		for i2 := 0; i2 < int(tabs); i2++ {
			tabulator = strings.Join([]string{tabulator, "\t"}, "")
		}
		splittedAddr := strings.Split(ipMap[i], ".") // Separating CIDR mask if present
		splittedOctet := strings.Split(splittedAddr[3], "/")
		if len(splittedOctet) == 1 {
			ipAddr = strings.Join(splittedAddr, ".")
		} else {
			splittedAddr[3] = splittedOctet[0]
			ipAddr = strings.Join(splittedAddr, ".")
		}
		fmt.Println(strings.Join([]string{i, tabulator, ipAddr}, ""))
	}
	fmt.Fprintln(os.Stderr, "Done!")
}

func detectPveCluster() {
	_, err := os.Stat("/etc/pve/corosync.conf")
	if err != nil {
		fmt.Println("Proxmox corosync config not detected on host. Exiting...")
		os.Exit(1)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
