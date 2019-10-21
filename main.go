package main

import (
	"bufio"
	"bytes"
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
	format := flag.String("format", "hosts", "output in ansible hosts definition format")
	flag.Parse()
	checkFormat(*format)
	detectPveCluster()
	fmt.Fprintln(os.Stderr, "Gathering info about cluster nodes...")
	list := getNodeNamelist()
	ipAddr := getIPaddressess(list, *iface)
	fmt.Fprintln(os.Stderr, "Printing result...")
	printResult(ipAddr, *format)
	fmt.Fprintln(os.Stderr, "Done!")
}

// Accessing Proxmox API to get list of nodes

func getNodeNamelist() []string {
	var list []string
	var v interface{}
	jsonRaw := apiGetReq("nodes")
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

//GET request to PVE API
func apiGetReq(path string) []byte {
	var rawJSON []byte
	resp, err := exec.Command("pvesh", "get", path, "-output-format", "json").CombinedOutput()
	check(err)
	reader := bytes.NewReader(resp)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		firstChar := string(scanner.Text()[0])
		if firstChar == "[" {
			rawJSON = []byte(scanner.Text())
			if json.Valid(rawJSON) {
				break
			} else {
				rawJSON = nil
			}
		}
	}
	return rawJSON
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
		rawJSON := apiGetReq(apiPath)
		if rawJSON == nil {
			fmt.Println("Failed to get information about", nodeName)
			os.Exit(10)
		}
		json.Unmarshal(rawJSON, &v)
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
	for k := range ipMap {
		splittedAddr := strings.Split(ipMap[k], ".") // Separating CIDR mask if present
		splittedOctet := strings.Split(splittedAddr[3], "/")
		if len(splittedOctet) == 1 {
			ipMap[k] = strings.Join(splittedAddr, ".")
		} else {
			splittedAddr[3] = splittedOctet[0]
			ipMap[k] = strings.Join(splittedAddr, ".")
		}
	}
	return ipMap
}

func printResult(ipMap map[string]string, format string) {
	var longest int
	longest = 0
	for k := range ipMap {
		if len(ipMap[k]) > longest {
			longest = len(ipMap[k])
		}
	}
	maxDivider, _ := math.Modf(float64(longest) / 8)
	for k := range ipMap {
		tabulator := "\t"
		divider, _ := math.Modf(float64(len(k)) / 8)
		tabs := maxDivider - divider
		for i := 0; i < int(tabs); i++ {
			tabulator = strings.Join([]string{tabulator, "\t"}, "")
		}
		if format == "ansible" {
			fmt.Printf("%s%sansible_host=%s\n", k, tabulator, ipMap[k])
		} else if format == "hosts" {
			fmt.Println(strings.Join([]string{ipMap[k], tabulator, k}, ""))
		}
	}
}

func printResultAnsible(ipMap map[string]string) {
	for k, v := range ipMap {
		splittedAddr := strings.Split(v, ".")
		splittedOctet := strings.Split(splittedAddr[3], "/")
		if len(splittedOctet) > 1 {
			splittedAddr[3] = splittedOctet[0]
		}
		v = strings.Join(splittedAddr, ".")
		fmt.Printf("%s ansible_host=%s\n", k, v)
	}
}

func detectPveCluster() {
	_, err := os.Stat("/etc/pve/corosync.conf")
	if err != nil {
		fmt.Println("Proxmox corosync config not detected on host. Exiting...")
		os.Exit(1)
	}
}

func checkFormat(format string) {
	switch format {
	case "ansible", "hosts":
		return
	default:
		fmt.Println("Incorrect format definition. Use ansible or hosts.")
		os.Exit(10)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
