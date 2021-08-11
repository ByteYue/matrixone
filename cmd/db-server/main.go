package main

import (
	"flag"
	"fmt"
	"github.com/fagongzi/log"
	"github.com/matrixorigin/matrixcube/components/prophet/util"
	"matrixone/pkg/config"
	"matrixone/pkg/frontend"
	"matrixone/pkg/logger"
	"matrixone/pkg/rpcserver"
	"matrixone/pkg/sql/handler"
	"matrixone/pkg/util/signal"
	aoe_catalog "matrixone/pkg/vm/engine/aoe/catalog"
	aoe_engine "matrixone/pkg/vm/engine/aoe/engine"
	"matrixone/pkg/vm/mempool"
	"matrixone/pkg/vm/metadata"
	"matrixone/pkg/vm/mmu/guest"
	"matrixone/pkg/vm/mmu/host"
	"matrixone/pkg/vm/process"
	"os"
)

var (
	mo   *frontend.MOServer
	pcis []*frontend.PDCallbackImpl
)

func createMOServer(callback *frontend.PDCallbackImpl) {
	address := fmt.Sprintf("%s:%d", config.GlobalSystemVariables.GetHost(), config.GlobalSystemVariables.GetPort())
	pu := config.NewParameterUnit(&config.GlobalSystemVariables, config.HostMmu, config.Mempool, config.StorageEngine, config.ClusterNodes)
	mo = frontend.NewMOServer(address, pu, callback)
}

func runMOServer() error {
	return mo.Start()
}

func serverShutdown(isgraceful bool) {
	mo.Stop()
}

func registerSignalHandlers() {
	signal.SetupSignalHandler(serverShutdown)
}

func cleanup() {
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s configFile\n", os.Args[0])
		os.Exit(-1)
	}
	flag.Parse()

	//close cube print info
	log.SetLevelByString("error")
	log.SetHighlighting(false)
	util.SetLogger(log.NewLoggerWithPrefix("prophet"))

	//before anything using the configuration
	if err := config.GlobalSystemVariables.LoadInitialValues(); err != nil {
		fmt.Printf("error:%v\n",err)
		return
	}

	if err := config.LoadvarsConfigFromFile(os.Args[1], &config.GlobalSystemVariables); err != nil {
		fmt.Printf("error:%v\n",err)
		return
	}

	fmt.Println("Shutdown The Server With Ctrl+C | Ctrl+\\.")

	config.HostMmu = host.New(config.GlobalSystemVariables.GetHostMmuLimitation())
	config.Mempool = mempool.New(int(config.GlobalSystemVariables.GetMempoolMaxSize()), int(config.GlobalSystemVariables.GetMempoolFactor()))

	if ! config.GlobalSystemVariables.GetDumpEnv() {
		fmt.Println("Using AOE Storage Engine, 3 Cluster Nodes, 1 SQL Server.")

		nodeCnt := 3
		pcis = make([]*frontend.PDCallbackImpl, nodeCnt)

		ppu := frontend.NewPDCallbackParameterUnit(
			int(config.GlobalSystemVariables.GetPeriodOfEpochTimer()),
			int(config.GlobalSystemVariables.GetPeriodOfPersistence()),
			int(config.GlobalSystemVariables.GetPeriodOfDDLDeleteTimer()),
			int(config.GlobalSystemVariables.GetTimeoutOfHeartbeat()))

		for i := 0 ; i < nodeCnt; i++ {
			pcis[i] = frontend.NewPDCallbackImpl(ppu)
			pcis[i].Id = i
		}

		c, err := frontend.NewTestClusterStore(nil,true,nil, pcis, nodeCnt)
		if err != nil {
			os.Exit(-2)
		}

		catalog := aoe_catalog.DefaultCatalog(c.Applications[0])
		eng := aoe_engine.Mock(&catalog)

		for i := 0 ; i < nodeCnt; i++ {
			pcis[i].SetCatalogService(&catalog)
		}

		//one rpcserver per cube node
		for i := 0 ; i < nodeCnt ; i++ {
			//db := c.AOEDBs[i].DB
			hm := config.HostMmu
			gm := guest.New(1<<40, hm)
			proc := process.New(gm, config.Mempool)
			{
				proc.Id = "0"
				proc.Lim.Size = 10 << 32
				proc.Lim.BatchRows = 10 << 32
				proc.Lim.PartitionRows = 10 << 32
				proc.Refer = make(map[string]uint64)
			}
			log := logger.New(os.Stderr, fmt.Sprintf("rpc%v:", i))
			log.SetLevel(logger.WARN)
			srv, err := rpcserver.New(fmt.Sprintf("127.0.0.1:%v", 20000+i+100), 1<<30, log)
			if err != nil {
				log.Fatal(err)
			}
			hp := handler.New(eng, proc)
			srv.Register(hp.Process)
			go srv.Run()
		}

		//test storage engine
		config.StorageEngine = eng

		//test cluster nodes
		config.ClusterNodes = metadata.Nodes{}
	}else{
		panic("The Official Storage Engine and Cluster Nodes are in the developing.")

		//TODO:
		config.StorageEngine = nil

		config.ClusterNodes = nil
	}
	fmt.Println("Create MOServer")
	createMOServer(pcis[0])
	err := runMOServer()
	if err != nil {
		fmt.Println(err)
		return
	}
	//registerSignalHandlers()

	select {}
	cleanup()
	os.Exit(0)
}