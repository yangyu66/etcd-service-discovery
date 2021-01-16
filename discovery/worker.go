/*
./etcdctl watch --prefix workers
go run main.go -role master
go run main.go -role worker
./etcdctl put workers/ca '{"IP":"127.0.0.1","Name":"localhost2","CPU":14}'

*/
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	client "github.com/coreos/etcd/clientv3"
)

// WorkerInfo is the service register information to etcd
type WorkerInfo struct {
	IP   string
	Name string
	CPU  int
}

// Worker Node
type Worker struct {
	Name string
	IP   string
	API  *client.Client
}

// NewWorker method.
func NewWorker(name, IP string, endpoints []string) *Worker {

	// etcd 配置
	cfg := client.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	// 创建 etcd 客户端
	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Fatal("Error: cannot connect to etcd: ", err)
	}

	// 创建 worker
	worker := &Worker{
		Name: name,
		IP:   IP,
		API:  etcdClient,
	}

	return worker
}

func (worker *Worker) HeartBeat() {

	for {

		// worker info
		info := &WorkerInfo{
			Name: worker.Name,
			IP:   worker.IP,
			CPU:  runtime.NumCPU(),
		}

		key := "workers/" + worker.Name
		value, _ := json.Marshal(info)
		fmt.Println(string(value))

		// 创建 lease
		leaseResp, err := worker.API.Grant(context.TODO(), 10)
		if err != nil {
			log.Fatalf("设置租约时间失败:%s\n", err.Error())
		}
		// 创建 watcher channel
		_, err = worker.API.Put(context.TODO(), key, string(value), client.WithLease(leaseResp.ID))
		if err != nil {
			log.Println("Error update workerInfo:", err)
		}
		keepAlive, err := worker.API.KeepAlive(context.TODO(), leaseResp.ID)
		for range keepAlive {
			// eat messages until keep alive channel closes
		}

		// time.Sleep(time.Second * 3)
	}
}
