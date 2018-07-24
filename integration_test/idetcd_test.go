//etcd local instance is needed.
package test

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/proxy"
	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"
	"github.com/mholt/caddy"
	"github.com/miekg/dns"
)

var localIP = getLocalIPAddress().String()

func TestBasicLookupNodesRR(t *testing.T) {
	corefiles := generateCorefiles(5)
	var udps []string
	var nodes []*caddy.Instance

	for i, corefile := range corefiles {
		node, udp, _, err := CoreDNSServerAndPorts(corefile)
		if err != nil {
			t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, i)
		}
		nodes = append(nodes, node)
		udps = append(udps, udp)
		defer node.Stop()
	}

	p := proxy.NewLookup([]string{udps[0]}) // use udp port from the server
	state := request.Request{W: &test.ResponseWriter{}, Req: new(dns.Msg)}

	for i := range corefiles {
		resp, err := p.Lookup(state, "worker"+strconv.Itoa(i+1)+".tf.local.", dns.TypeA)
		if err != nil {
			t.Fatalf("Expected to receive reply, but didn't: %v", err)
		}
		if len(resp.Answer) == 0 {
			t.Fatalf("Expected to at least one RR in the answer section, got none")
		}
		if resp.Answer[0].Header().Rrtype != dns.TypeA {
			t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
		}
		if resp.Answer[0].(*dns.A).A.String() != localIP {
			t.Errorf("Expected %s , got: %s", localIP, resp.Answer[0].(*dns.A).A.String())
		}
	}
	time.Sleep(10 * time.Second)
	for _, node := range nodes {
		node.ShutdownCallbacks()
	}
	time.Sleep(30 * time.Second)
}

func TestNodeUpAfterTTL(t *testing.T) {
	corefiles := generateCorefiles(5)
	var udps []string
	var nodes []*caddy.Instance

	for i, corefile := range corefiles[:4] {
		node, udp, _, err := CoreDNSServerAndPorts(corefile)
		if err != nil {
			t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, i)
		}
		nodes = append(nodes, node)
		udps = append(udps, udp)
		defer node.Stop()
	}

	p := proxy.NewLookup([]string{udps[0]}) // use udp port from the server
	state := request.Request{W: &test.ResponseWriter{}, Req: new(dns.Msg)}

	for i := range corefiles[:4] {
		resp, err := p.Lookup(state, "worker"+strconv.Itoa(i+1)+".tf.local.", dns.TypeA)
		if err != nil {
			t.Fatalf("Expected to receive reply, but didn't: %v", err)
		}
		if len(resp.Answer) == 0 {
			t.Fatalf("Expected to at least one RR in the answer section, got none")
		}
		if resp.Answer[0].Header().Rrtype != dns.TypeA {
			t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
		}
		if resp.Answer[0].(*dns.A).A.String() != localIP {
			t.Errorf("Expected %s , got: %s", localIP, resp.Answer[0].(*dns.A).A.String())
		}
	}
	////should fail when check the 5th one
	resp, err := p.Lookup(state, "worker"+strconv.Itoa(5)+".tf.local.", dns.TypeA)
	if err != nil {
		t.Fatalf("Expected to receive reply, but didn't: %v", err)
	}
	if len(resp.Answer) != 0 {
		t.Fatalf("Shouldn't have the RR for node 5!")
	}
	time.Sleep(20 * time.Second)

	//now node 5 is up
	node5, _, _, err := CoreDNSServerAndPorts(corefiles[4])
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, 5)
	}
	nodes = append(nodes, node5)
	defer node5.Stop()
	//check rr of node 5
	resp, err = p.Lookup(state, "worker"+strconv.Itoa(5)+".tf.local.", dns.TypeA)
	if err != nil {
		t.Fatalf("Expected to receive reply, but didn't: %v", err)
	}
	if len(resp.Answer) == 0 {
		t.Fatalf("Expected to at least one RR in the answer section, got none")
	}
	if resp.Answer[0].Header().Rrtype != dns.TypeA {
		t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
	}
	if resp.Answer[0].(*dns.A).A.String() != localIP {
		t.Errorf("Expected %s , got: %s", localIP, resp.Answer[0].(*dns.A).A.String())
	}
	for _, node := range nodes {
		node.ShutdownCallbacks()
	}
	time.Sleep(25 * time.Second)
}

func TestNodeTakeFreeSlot(t *testing.T) {
	corefiles := generateCorefiles(5)
	shutDownIdx := 1
	var udps []string
	var nodes []*caddy.Instance

	for i, corefile := range corefiles[:4] {
		node, udp, _, err := CoreDNSServerAndPorts(corefile)
		if err != nil {
			t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, i)
		}
		nodes = append(nodes, node)
		udps = append(udps, udp)
		if i != shutDownIdx {
			defer node.Stop()
		}
	}

	p := proxy.NewLookup([]string{udps[0]}) // use udp port from the server
	state := request.Request{W: &test.ResponseWriter{}, Req: new(dns.Msg)}

	for i := range corefiles[:4] {
		resp, err := p.Lookup(state, "worker"+strconv.Itoa(i+1)+".tf.local.", dns.TypeA)
		if err != nil {
			t.Fatalf("Expected to receive reply, but didn't: %v", err)
		}
		if len(resp.Answer) == 0 {
			t.Fatalf("Expected to at least one RR in the answer section, got none")
		}
		if resp.Answer[0].Header().Rrtype != dns.TypeA {
			t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
		}
		if resp.Answer[0].(*dns.A).A.String() != localIP {
			t.Errorf("Expected %s , got: %s", localIP, resp.Answer[0].(*dns.A).A.String())
		}
	}
	time.Sleep(8 * time.Second)
	//shut down node 2
	nodes[shutDownIdx].ShutdownCallbacks()
	nodes[shutDownIdx].Stop()
	//Free the slot of node 2
	time.Sleep(20 * time.Second)
	resp, err := p.Lookup(state, "worker"+strconv.Itoa(2)+".tf.local.", dns.TypeA)
	if err != nil {
		t.Fatalf("Expected to receive reply, but didn't: %v", err)
	}
	if len(resp.Answer) != 0 {
		t.Fatalf("Shouldn't have the RR for node 2!")
	}

	//Ask noede 5 to take the free slot
	//now node 5 is up
	node5, _, _, err := CoreDNSServerAndPorts(corefiles[4])
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, 5)
	}
	nodes = append(nodes, node5)
	defer node5.Stop()
	//check rr of node 5
	resp, err = p.Lookup(state, "worker"+strconv.Itoa(2)+".tf.local.", dns.TypeA)
	if err != nil {
		t.Fatalf("Expected to receive reply, but didn't: %v", err)
	}
	if len(resp.Answer) == 0 {
		t.Fatalf("Expected to at least one RR in the answer section, got none")
	}
	if resp.Answer[0].Header().Rrtype != dns.TypeA {
		t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
	}
	if resp.Answer[0].(*dns.A).A.String() != localIP {
		t.Errorf("Expected %s , got: %s", localIP, resp.Answer[0].(*dns.A).A.String())
	}
	node2, _, _, err := CoreDNSServerAndPorts(corefiles[shutDownIdx])
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, shutDownIdx)
	}
	nodes[shutDownIdx] = node2
	defer node2.Stop()
	//check rr of node 2
	resp, err = p.Lookup(state, "worker"+strconv.Itoa(5)+".tf.local.", dns.TypeA)
	if err != nil {
		t.Fatalf("Expected to receive reply, but didn't: %v", err)
	}
	if len(resp.Answer) == 0 {
		t.Fatalf("Expected to at least one RR in the answer section, got none")
	}
	if resp.Answer[0].Header().Rrtype != dns.TypeA {
		t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
	}
	if resp.Answer[0].(*dns.A).A.String() != localIP {
		t.Errorf("Expected %s , got: %s", localIP, resp.Answer[0].(*dns.A).A.String())
	}

	for _, node := range nodes {
		node.ShutdownCallbacks()
	}
	time.Sleep(20 * time.Second)
}

func generateCorefiles(numNode int) []string {
	var corefiles []string
	limit := strconv.Itoa(numNode)
	for i := 0; i < numNode; i++ {
		port := strconv.Itoa(1053 + i)
		corefile := `.:` + port + ` {
			idetcd {
				endpoint http://localhost:2379
				pattern worker{{.ID}}.tf.local.
				limit ` + limit + `
			}
		}`
		corefiles = append(corefiles, corefile)
	}
	return corefiles
}
func getLocalIPAddress() net.IP {
	var localIP net.IP
	interfaces, _ := net.Interfaces()
	for _, inter := range interfaces {
		addrs, _ := inter.Addrs()
		for _, addr := range addrs {
			localIP = net.ParseIP(strings.Split(addr.String(), "/")[0])
			if localIP.To4() != nil && !localIP.IsLoopback() {
				return localIP
			}
		}
	}
	return localIP
}
