package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yodo-io/cryptogen/pkg/api"
	"github.com/yodo-io/cryptogen/pkg/crypto/worker"

	"github.com/yodo-io/cryptogen/pkg/crypto"
	"github.com/yodo-io/cryptogen/pkg/kms"
	"github.com/yodo-io/cryptogen/pkg/store"
)

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Env var %s must be set", key)
	}
	return v
}

type config struct {
	address        string
	tmpDir         string
	vaultTokenPath string
	redisAddress   string
	cryptogenPath  string
}

func mustConfigure() config {
	c := config{}
	c.address = mustGetEnv("ADDRESS")
	c.tmpDir = mustGetEnv("TMP_DIR")
	c.vaultTokenPath = mustGetEnv("VAULT_TOKEN_PATH")
	c.redisAddress = mustGetEnv("REDIS_ADDR")
	c.cryptogenPath = mustGetEnv("CRYPTOGEN_PATH")
	return c
}

func main() {
	c := mustConfigure()
	signals()
	server(c)
}

func signals() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)
	go func() {
		sig := <-stop
		log.Printf("Caught %v", sig)
		os.Exit(0)
	}()
}

func server(c config) {
	// kms service
	kms, err := kms.NewVault(kms.VaultConf{
		TokenPath: c.vaultTokenPath,
	})
	if err != nil {
		log.Fatal(err)
	}

	// crypto service and worker
	cg, err := crypto.New(crypto.Config{
		TmpDir:        c.tmpDir,
		CryptogenPath: c.cryptogenPath,
	})
	if err != nil {
		log.Fatal(err)
	}
	w := worker.New(&worker.Config{
		Kms:    kms,
		Crypto: cg,
	})

	// data store
	s := store.NewRedis(store.RedisConf{
		Address: c.redisAddress,
	})
	// processes job updates
	store.UpdateJobs(s, w.Feed)

	// api
	r := api.New(&api.Config{
		Worker: w,
		Store:  s,
	})
	r.Run(c.address)
}
