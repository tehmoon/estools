package main

import (
	"./rules"
	"./scheduler"
	"./client"
	"./storage"
	"./alert"
	"./api"
	"./flags"
	"./util"
	"./response"
	"github.com/olivere/elastic"
)

func main() {
	f, err := flags.Parse()
	if err != nil {
		util.Fatal(err)
	}

	rm := rules.NewManager(&rules.ManagerConfig{
		Owners: f.Owners,
		Exec: f.Exec,
		Index: f.Index,
	})

	err = rm.LoadRules(f.Dir)
	if err != nil {
		util.Fatal(err)
	}

	cm, err := client.NewManager(&client.ManagerConfig{
		ElasticConfigs: []elastic.ClientOptionFunc{
			elastic.SetSniff(false),
			elastic.SetURL(f.Server),
		},
	})
	if err != nil {
		util.Fatal(err)
	}

	stm, err := storage.NewManager(&storage.ManagerConfig{
		StorageName: "memory",
		ClientManager: cm,
	})
	if err != nil {
		util.Fatal(err)
	}

	rsm, err := response.NewManager(&response.ManagerConfig{})
	if err != nil {
		util.Fatal(err)
	}

	am := alert.NewManager(&alert.ManagerConfig{
		PublicURL: f.PublicURL,
		Exec: f.Exec,
		StorageManager: stm,
		ResponseManager: rsm,
	})

	apm, err := api.NewManager(&api.ManagerConfig{
		Listen: f.Listen,
		StorageManager: stm,
		AlertManager: am,
		ResponseManager: rsm,
	})
	if err != nil {
		util.Fatal(err)
	}

	scheduler.NewManager(&scheduler.ManagerConfig{
		QueryDelay: f.QueryDelay,
		RulesManager: rm,
		ClientManager: cm,
		AlertManager: am,
	})

	err = apm.Start()
	if err != nil {
		util.Fatal(err)
	}
}
