//Package idetcd implements a plugin that allow host confiuration config itself without collison by using etcd.
package idetcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"text/template"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	etcdcv3 "github.com/coreos/etcd/clientv3"
	"github.com/miekg/dns"
)

//Idetcd is a plugin which can configure the cluster without collison.
type Idetcd struct {
	Next      plugin.Handler
	Ctx       context.Context
	Client    *etcdcv3.Client
	endpoints []string
	pattern   *template.Template
	ID        int
	limit     int
}

//ServeDNS implements the plugin.Handler interface
func (idetcd *Idetcd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true
	qname := state.Name()
	fmt.Println(qname)
	resp, _ := idetcd.get(qname)
	record := new(Record)
	if err := json.Unmarshal(resp.Kvs[0].Value, record); err != nil {
		return plugin.NextOrFailure(idetcd.Name(), idetcd.Next, ctx, w, r)
	}
	fmt.Printf("%v\n", record)
	var rr dns.RR
	switch state.QType() {
	case dns.TypeA:
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: state.QClass()}
		rr.(*dns.A).A = net.ParseIP(record.Ipv4).To4()
	case dns.TypeAAAA:
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeAAAA, Class: state.QClass()}
		rr.(*dns.AAAA).AAAA = net.ParseIP(record.Ipv6).To16()

	}
	a.Answer = []dns.RR{rr}
	w.WriteMsg(a)
	return plugin.NextOrFailure(idetcd.Name(), idetcd.Next, ctx, w, r)
}

//set is a wrapper for client.Set
func (idetcd *Idetcd) set(key string, value string, opts ...etcdcv3.OpOption) (*etcdcv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(idetcd.Ctx, 5*time.Second)
	defer cancel()
	r, err := idetcd.Client.Put(ctx, key, value, opts...)
	if err != nil {
		return r, err
	}
	return r, nil
}

// get is a wrapper for client.Get
func (idetcd *Idetcd) get(key string) (*etcdcv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(idetcd.Ctx, 5*time.Second)
	defer cancel()
	r, err := idetcd.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//Name implements the Handler interface.
func (idetcd *Idetcd) Name() string { return "idetcd" }
