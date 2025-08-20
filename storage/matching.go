package storage

// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/topfreegames/pitaya/v3/pkg/cluster"
	"github.com/topfreegames/pitaya/v3/pkg/config"
	"github.com/topfreegames/pitaya/v3/pkg/constants"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"github.com/topfreegames/pitaya/v3/pkg/modules"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
)

type Matching struct {
	ServerId   string `json:"server_id"`
	ServerType string `json:"server_type"`
}

// ETCDMatching module that uses etcd to keep in which frontend server each user is bound
type ETCDMatching struct {
	modules.Base
	cli             *clientv3.Client
	etcdEndpoints   []string
	etcdPrefix      string
	etcdDialTimeout time.Duration
	leaseTTL        time.Duration
	leaseID         clientv3.LeaseID
	thisServer      *cluster.Server
	stopChan        chan struct{}
}

// NewETCDMatching returns a new instance of BindingStorage
func NewETCDMatching(server *cluster.Server, conf config.ETCDBindingConfig) *ETCDMatching {
	b := &ETCDMatching{
		thisServer: server,
		stopChan:   make(chan struct{}),
	}
	b.etcdDialTimeout = conf.DialTimeout
	b.etcdEndpoints = conf.Endpoints
	b.etcdPrefix = conf.Prefix
	b.leaseTTL = conf.LeaseTTL
	return b
}

func getUserMatchingKey(uid string) string {
	return fmt.Sprintf("matching/%s", uid)
}

// Put puts the binding info into etcd
func (b *ETCDMatching) Put(uid string) error {
	matching := Matching{
		ServerId:   b.thisServer.ID,
		ServerType: b.thisServer.Type,
	}

	value, err := json.Marshal(matching)
	if err != nil {
		return err
	}
	_, err = b.cli.Put(context.Background(), getUserMatchingKey(uid), string(value), clientv3.WithLease(b.leaseID))
	return err
}

func (b *ETCDMatching) Remove(uid string) error {
	_, err := b.cli.Delete(context.Background(), getUserMatchingKey(uid))
	return err
}

// Get gets the id of the match server a user is connected to
// TODO: should we set context here?
// TODO: this could be way more optimized, using watcher and local caching
func (b *ETCDMatching) Get(uid string) (*Matching, error) {
	etcdRes, err := b.cli.Get(context.Background(), getUserMatchingKey(uid))
	if err != nil {
		return nil, err
	}
	if len(etcdRes.Kvs) == 0 {
		return nil, constants.ErrBindingNotFound
	}

	matching := &Matching{}
	err = json.Unmarshal(etcdRes.Kvs[0].Value, matching)
	return matching, err
}

func (b *ETCDMatching) watchLeaseChan(c <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case <-b.stopChan:
			return
		case kaRes := <-c:
			if kaRes == nil {
				logger.Log.Warn("[binding storage] sd: error renewing etcd lease, rebootstrapping")
				for {
					err := b.bootstrapLease()
					if err != nil {
						logger.Log.Warn("[binding storage] sd: error rebootstrapping lease, will retry in 5 seconds")
						time.Sleep(5 * time.Second)
						continue
					} else {
						return
					}
				}
			}
		}
	}
}

func (b *ETCDMatching) bootstrapLease() error {
	// grab lease
	l, err := b.cli.Grant(context.TODO(), int64(b.leaseTTL.Seconds()))
	if err != nil {
		return err
	}
	b.leaseID = l.ID
	logger.Log.Debugf("[binding storage] sd: got leaseID: %x", l.ID)
	// this will keep alive forever, when channel c is closed
	// it means we probably have to rebootstrap the lease
	c, err := b.cli.KeepAlive(context.TODO(), b.leaseID)
	if err != nil {
		return err
	}
	// need to receive here as per etcd docs
	<-c
	go b.watchLeaseChan(c)
	return nil
}

// Init starts the binding storage module
func (b *ETCDMatching) Init() error {
	var cli *clientv3.Client
	var err error
	if b.cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   b.etcdEndpoints,
			DialTimeout: b.etcdDialTimeout,
		})
		if err != nil {
			return err
		}
		b.cli = cli
	}
	// namespaced etcd :)
	b.cli.KV = namespace.NewKV(b.cli.KV, b.etcdPrefix)
	err = b.bootstrapLease()
	if err != nil {
		return err
	}

	return nil
}

// Shutdown executes on shutdown and will clean etcd
func (b *ETCDMatching) Shutdown() error {
	close(b.stopChan)
	return b.cli.Close()
}
