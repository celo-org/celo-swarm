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

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
)

var (
	loglevel = flag.Int("loglevel", 5, "verbosity of logs")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

/*
test case:
1. swarm-to-swarm connection
2. swarm-to-eth node



*/

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
					Msg: &Handshake{
						ServeHeaders: true,
					},
					Peer: node.ID(),
				},
			},
			Expects: []p2ptest.Expect{
				{
					Code: 0,
					Msg: &Handshake{
						ServeHeaders: false,
					},
					Peer: node.ID(),
				},
			},
		})

	if err != nil {
		t.Fatalf("Got %v", err)
	}

	close(b.quit)

	err = tester.TestDisconnected()
	if err != nil {
		t.Fatal(err)
	}
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
						ServeHeaders: false,
					},
					Peer: node.ID(),
				},
			},
		})

	if err != nil {
		t.Fatalf("Got %v", err)
	}

	close(b.quit)

	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("protocol returned")})
	if err != nil {
		t.Fatal(err)
	}
}

//func TestNodesCanTalk(t *testing.T) {
//nodeCount := 2

//// create a standard sim
//sim := simulation.NewInProc(map[string]simulation.ServiceFunc{
//"bzz-eth": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
//addr := network.NewAddr(ctx.Config.Node())

//o := New(nil)
//cleanup = func() {
//}

//return o, cleanup, nil
//},
//})
//defer sim.Close()

//ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
//defer cancel()
//_, err := sim.AddNodesAndConnectStar(nodeCount)
//if err != nil {
//t.Fatal(err)
//}

////run the simulation
//result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
//log.Info("Simulation running")
//_ = sim.Net.Nodes

////wait until all subscriptions are done
//select {
//case <-ctx.Done():
//return errors.New("Context timed out")
//}

//return nil
//})
//if result.Error != nil {
//t.Fatal(result.Error)
//}
//}
