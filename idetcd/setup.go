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

	//parse the corefile.
	idetc, err := idetcdParse(c)

	if err != nil {
		return plugin.Error("idetcd", err)
	}
	if c.NextArg() {
		return plugin.Error("idetcd", c.ArgErr())
	}

	//killChan is a channel used for integration tests.
	var (
		namebuf  bytes.Buffer
		killChan chan struct{}
		id       = 1
	)

	//get ipv4, ipv6 and port.
	host := iP()
	host.Port = dnsserver.GetConfig(c).Port
	fmt.Println(host)

	//put them in json format.
	localIP, err := json.Marshal(host)
	if err != nil {
		return plugin.Error("idetcd", err)
	}
	value := string(localIP)
	killChan = make(chan struct{})

	//Try to find a free slot for current node
	for id <= idetc.limit {
		idetc.ID = id
		idetc.pattern.Execute(&namebuf, idetc)
		name := namebuf.String()

		//Try to see if it can find the proposed domain name in the etcd
		resp, err := idetc.get(name)
		if err != nil {
			return err
		}

		if resp.Count == 0 {
			//Can not find the proposed domain name in the etcd, so the node take this domain name, and put the record attached a lease with ttl in etcd.
			lease, _ := idetc.Client.Grant(context.TODO(), defaultTTL)
			idetc.set(name, value, etcdcv3.WithLease(lease.ID))
			break
		} else {
			//Node find the proposed domain name is already used by other node, so it increases the proposed id and try another domain name.
			id++
			namebuf.Reset()
		}
	}

	//If node can not find a free slot until it proposed id is bigger than the limit, then just stop the coredns server.
	if id > idetc.limit {
		return plugin.Error("idetcd", c.Errf("Could not have more than %d nodes in you cluster.", idetc.limit))
	}

	//update the record in the etcd
	//Here node renew its own record in etcd periodly.Every time the node want to renew record, it should first get its previous record in etcd, and compare it with
	//the record it has locally, if they are the same, then the node update it.
	//Notice that the period is smaller than ttl, this is because if the period is exactly the ttl, sometimes the record can be deleted by the etcd before node
	//updates since communicating with etcd also needs some time.
	renewTicker := time.NewTicker(defaultTTL / 2 * time.Second)
	go func() {
		for {
			select {
			case <-renewTicker.C:
				resp, err := idetc.get(namebuf.String())
				if err != nil {
					return
				}
				if string(resp.Kvs[0].Value) == value {
					lease, _ := idetc.Client.Grant(context.TODO(), defaultTTL)
					idetc.set(namebuf.String(), value, etcdcv3.WithLease(lease.ID))
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

//Get the both ipv4 and ipv6 local address(in Record format) of the interface which is the first one after loopback interface.
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

//Parsing the Corefile.
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

//Return a etcd client.
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
