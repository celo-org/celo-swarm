// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package bzzeth

import (
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
)

var (
	loglevel = flag.Int("loglevel", 3, "verbosity of logs")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

func newBzzEthTester() (*p2ptest.ProtocolTester, *BzzEth, func(), error) {
	b := New(nil)

	prvkey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, nil, err
	}

	protocolTester := p2ptest.NewProtocolTester(prvkey, 1, b.Run)
	teardown := func() {
		protocolTester.Stop()
	}

	return protocolTester, b, teardown, nil
}

// tests handshake between eth node and swarm node
// on successful handshake the protocol does not go idle
// and serves headers is registered
func TestBzzEthHandshake(t *testing.T) {
	tester, b, teardown, err := newBzzEthTester()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	node := tester.Nodes[0]

	err = tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Handshake",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: Handshake{
						ServeHeaders: true,
					},
					Peer: node.ID(),
				},
			},
			Expects: []p2ptest.Expect{
				{
					Code: 0,
					Msg: Handshake{
						ServeHeaders: true,
					},
					Peer: node.ID(),
				},
			},
		})

	if err != nil {
		t.Fatalf("Got %v", err)
	}
	var p *Peer
	for i := 0; i < 10; i++ {
		p = b.getPeer(node.ID())
		if p != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if p == nil {
		t.Fatal("bzzeth peer not added")
	}
	if !p.serveHeaders {
		t.Fatal("bzzeth peer serveHeaders not set")
	}

	close(b.quit)

	err = tester.TestDisconnected()
	if err != nil {
		t.Fatal(err)
	}
}

// TestBzzBzzHandshake tests that a handshake between two Swarm nodes
func TestBzzBzzHandshake(t *testing.T) {
	tester, b, teardown, err := newBzzEthTester()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	defer func(f func(*Peer) bool) {
		isSwarmNodeFunc = f
	}(isSwarmNodeFunc)

	isSwarmNodeFunc = func(_ *Peer) bool { return true }
	node := tester.Nodes[0]

	err = tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Handshake",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: &Handshake{
						ServeHeaders: false,
					},
					Peer: node.ID(),
				},
			},
			Expects: []p2ptest.Expect{
				{
					Code: 0,
					Msg: &Handshake{
						ServeHeaders: true,
					},
					Peer: node.ID(),
				},
			},
		})

	if err != nil {
		t.Fatalf("Got %v", err)
	}

	p := b.getPeer(node.ID())
	if p != nil {
		t.Fatal("bzzeth swarm peer incorrectly added")
	}

	close(b.quit)

	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("protocol returned")})
	if err != nil {
		t.Fatal(err)
	}

}