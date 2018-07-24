package idetcd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	etcdcv3 "github.com/coreos/etcd/clientv3"

	"github.com/mholt/caddy"
)

const (
	defaultEndpoint = "http://localhost:2379"
	defaultTTL      = 20
	defaultLimit    = 10
)

func init() {
	caddy.RegisterPlugin("idetcd", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	idetc, err := idetcdParse(c)
	if err != nil {
		return plugin.Error("idetcd", err)
	}
	if c.NextArg() {
		return plugin.Error("idetcd", c.ArgErr())
	}

	var (
		namebuf  bytes.Buffer
		killChan chan struct{}
		id       = 1
	)

	host := iP()
	host.Port = dnsserver.GetConfig(c).Port
	localIP, err := json.Marshal(host)
	value := string(localIP)
	fmt.Printf("%v\n", value)
	if err != nil {
		return plugin.Error("idetcd", err)
	}
	killChan = make(chan struct{})

	for id <= idetc.limit {
		idetc.ID = id
		idetc.pattern.Execute(&namebuf, idetc)
		name := namebuf.String()
		resp, err := idetc.get(name)
		if err != nil {
			return err
		}
		if resp.Count == 0 {
			lease, err := idetc.Client.Grant(context.TODO(), defaultTTL)
			_, err = idetc.set(name, value, etcdcv3.WithLease(lease.ID))
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Printf("set node %s with ip %s\n", name, value)
			break
		} else {
			fmt.Printf("node %s is already exist!\n", name)
			id++
			namebuf.Reset()
		}
	}

	if id > idetc.limit {
		return plugin.Error("idetcd", c.Errf("Could not have more than %d nodes in you cluster.", idetc.limit))
	}

	//update the record in the etcd
	renewTicker := time.NewTicker(defaultTTL * time.Second)
	go func() {
		for {
			select {
			case <-renewTicker.C:
				resp, err := idetc.get(namebuf.String())
				if err != nil {
					return
				}
				if resp.Count == 0 || string(resp.Kvs[0].Value) == value {
					lease, _ := idetc.Client.Grant(context.TODO(), defaultTTL)
					idetc.set(namebuf.String(), value, etcdcv3.WithLease(lease.ID))
					fmt.Printf("Renew node %s with ip: %s\n", namebuf.String(), value)
				}
			case <-killChan:
				namebuf.Reset()
				return
			}
		}
	}()

	c.OnShutdown(func() error {
		close(killChan)
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		idetc.Next = next
		return idetc
	})
	return nil
}

func iP() Record {
	record := new(Record)
	interfaces, _ := net.Interfaces()
	var flag bool
	for _, inter := range interfaces {
		if inter.Flags&net.FlagLoopback == 0 {
			flag = false
			addrs, _ := inter.Addrs()
			for _, addr := range addrs {
				localIP := net.ParseIP(strings.Split(addr.String(), "/")[0])
				if localIP.To4() != nil {
					record.Ipv4 = localIP.String()
					flag = true
				} else if localIP.To16() != nil {
					record.Ipv6 = localIP.To16().String()
				}
			}
			if flag {
				break
			}
		}
	}
	return *record
}

func idetcdParse(c *caddy.Controller) (*Idetcd, error) {
	idetc := Idetcd{
		Ctx: context.Background(),
	}
	var (
		endpoints = []string{defaultEndpoint}
		pattern   = template.New("idetcd")
		limit     = defaultLimit
		err       error
	)
	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "endpoint":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return &Idetcd{}, c.ArgErr()
				}
				endpoints = args
			case "pattern":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return &Idetcd{}, c.ArgErr()
				}
				pattern, err = pattern.Parse(args[0])
				if err != nil {
					return &Idetcd{}, c.ArgErr()
				}
			case "limit":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return &Idetcd{}, c.ArgErr()
				}
				limit, err = strconv.Atoi(args[0])
				if err != nil {
					return &Idetcd{}, c.ArgErr()
				}
			}
		}
	}
	client, err := newEtcdClient(endpoints)
	if err != nil {
		return &Idetcd{}, err
	}
	idetc.endpoints = endpoints
	idetc.Client = client
	idetc.pattern = pattern
	idetc.limit = limit
	return &idetc, nil

}

func newEtcdClient(endpoints []string) (*etcdcv3.Client, error) {
	etcdCfg := etcdcv3.Config{
		Endpoints: endpoints,
	}
	cli, err := etcdcv3.New(etcdCfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
