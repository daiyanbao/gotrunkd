// gotrunkd project main.go
package main

import (
	//	"code.google.com/p/tuntap"
	"code.google.com/p/goconf/conf"
	"flag"
	"fmt"
	"log/syslog"
	"net"
)

func main() {
	//	var config *conf.ConfigFile
	var err error
	connectInfo := new(ConnectInfo)
	// comand line parse
	flag.BoolVar(&connectInfo.isServer, "server", false, "Run as server")
	flag.StringVar(&connectInfo.configPath, "config", "/etc/gotrunkd.default.config", "Path to config file")

	flag.IntVar(&connectInfo.port, "port", 5000, "Port number")

	flag.Parse()
	flag.Parsed()
	connectInfo.addr = net.ParseIP(flag.Arg(0)).String()

	connectInfo.logger, err = syslog.New(syslog.LOG_WARNING|syslog.LOG_INFO|syslog.LOG_DEBUG, "gotrunk")

	// config file parse
	c, err := conf.ReadConfigFile(connectInfo.configPath)
	checkError(err)
	connectInfo.logger.Info(fmt.Sprintf("Parse config:%s", connectInfo.configPath))

	connectInfo.port, err = c.GetInt("default", "port")
	checkError(err)
	connectInfo.protocol, err = c.GetString("default", "proto")
	checkError(err)
	connectInfo.logger.Debug(fmt.Sprintf("addr %s:%d proto %s", connectInfo.addr, connectInfo.port, connectInfo.protocol))
	sections := c.GetSections()
	connectInfo.logger.Debug(fmt.Sprintf("section num %d %s", len(sections), sections))
	connectInfo.routineNum = len(sections)
	connectInfo.trunkData = make(map[string]TrunkData)
	for _, value := range sections {
		if value == "default" {
			continue
		}
		tmpMap := connectInfo.trunkData[value] //todo hack for bug golang Issue 3117
		tmpMap.name = value
		tmpMap.password, err = c.GetString(value, "passwd")
		checkError(err)
		tmpMap.tun, err = c.GetString(value, "device")
		checkError(err)
		tmpMap.ipFrom, err = c.GetString(value, "ipFrom")
		checkError(err)
		tmpMap.ipTo, err = c.GetString(value, "ipTo")
		checkError(err)
		connectInfo.trunkData[value] = tmpMap
	}
	if connectInfo.isServer {
		connectInfo.title = "VTRUNKD server version 0.1go"
		server(connectInfo)
	} else {
		connectInfo.title = "VTRUNKD client version 0.1go"
		for key, _ := range connectInfo.trunkData {
			go client(connectInfo, key)
		}
	}
}
