package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type IPView struct {
	Ok        bool     `json:"ok"`
	Whitelist []string `json:"whitelist"`
	Blacklist []string `json:"blacklist"`
}

func main() {
	pflag.String("addr", "http://localhost:8888", "address of the antibruteforce service")
	pflag.Bool("reset-all", false, "reset all buckets")
	pflag.String("whitelist-add", "", "add IP(s) to whitelist (comma-separated)")
	pflag.String("whitelist-del", "", "remove IP(s) from whitelist (comma-separated)")
	pflag.String("blacklist-add", "", "add IP(s) to blacklist (comma-separated)")
	pflag.String("blacklist-del", "", "remove IP(s) from blacklist (comma-separated)")
	pflag.Bool("view-lists", false, "view whitelist and blacklist")

	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvPrefix("cli")
	viper.AutomaticEnv()

	addr := viper.GetString("addr")

	if viper.GetBool("reset-all") {
		resetBuckets(addr)
		return
	}

	if ips := viper.GetString("whitelist-add"); ips != "" {
		whitelistAdd(addr, ips)
		return
	}

	if ips := viper.GetString("whitelist-del"); ips != "" {
		whitelistDel(addr, ips)
		return
	}

	if ips := viper.GetString("blacklist-add"); ips != "" {
		blacklistAdd(addr, ips)
		return
	}

	if ips := viper.GetString("blacklist-del"); ips != "" {
		blacklistDel(addr, ips)
		return
	}

	if viper.GetBool("view-lists") {
		resp := doGet(addr + "/api/view/lists")
		var ipView IPView
		if err := json.Unmarshal(resp, &ipView); err != nil {
			log.Fatalf("failed to unmarshal response: %v", err)
		}
		if len(ipView.Whitelist) > 0 {
			fmt.Printf("Whitelisted: %s\n", strings.Join(ipView.Whitelist, ", "))
		} else {
			fmt.Printf("Whitelist is empty\n")
		}
		if len(ipView.Blacklist) > 0 {
			fmt.Printf("Blacklisted: %s\n", strings.Join(ipView.Blacklist, ", "))
			return
		}
		fmt.Printf("Blacklist is empty\n")
		return
	}
	pflag.PrintDefaults()
}
