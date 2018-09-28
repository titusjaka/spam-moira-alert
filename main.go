package main

import (
	"fmt"
	"time"

	"flag"
	"github.com/gosexy/to"
	"github.com/moira-alert/moira/cmd"
	"gopkg.in/fgrosse/graphigo.v2"
	"math/rand"
	"os"
)

var (
	configFileName = flag.String("config", "config.yml", "Path to configuration file")
)

type spamConfig struct {
	Nodes    int    `yaml:"nodes"`
	Metrics  int    `yaml:"metrics"`
	Interval string `yaml:"interval"`
	Address  string `yaml:"address"`
	Prefix   string `yaml:"prefix"`
	Values   int    `yaml:"values"`
}

func (conf *spamConfig) String() string {
	return fmt.Sprintf("Nodes: %d\nMetrics: %d\nInterval: %s\nAddress: %s\nPrefix: %s\n",
		conf.Nodes, conf.Metrics, conf.Interval, conf.Address, conf.Prefix)
}

func main() {
	flag.Parse()
	config := spamConfig{}

	err := cmd.ReadConfig(*configFileName, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't read settings: %v\n", err)
		os.Exit(1)
	}

	rand.Seed(time.Now().Unix())
	c := graphigo.Client{
		Address: config.Address,
		Timeout: 0,
		Prefix:  config.Prefix,
	}

	if err := c.Connect(); err != nil {
		panic(err)
	}

	defer c.Close()

	ticker := time.NewTicker(to.Duration(config.Interval))
	oldInterval := config.Interval
	configTicker := time.NewTicker(time.Second * 1)
	for {
		select {
		case y := <-ticker.C:
			sendMetrics(c, y, config.Nodes, config.Metrics, config.Values)
		case <-configTicker.C:

			cmd.ReadConfig(*configFileName, &config)
			if oldInterval != config.Interval {
				fmt.Printf("New config interval: %s\n", config.Interval)
				oldInterval = config.Interval
				ticker.Stop()
				ticker = time.NewTicker(to.Duration(config.Interval))
			}
		}
	}
}

func sendMetrics(c graphigo.Client, y time.Time, nodesCount, metricsCount, values int) {
	var metrics = make([]graphigo.Metric, 0)
	for i := 0; i < nodesCount; i++ {
		for j := 0; j < metricsCount; j++ {
			name := fmt.Sprintf("Metric_%v.%v", i, j)
			value := rand.Int() % values
			metrics = append(metrics, graphigo.Metric{Name: name, Value: value})
		}
	}
	c.SendAll(metrics)
	fmt.Printf("%v Metrics: %v\n", y.Format("2006-01-02 15:04:05"), len(metrics))

	for _, v := range metrics {
		fmt.Printf("\t%v.%v - %v\n", c.Prefix, v.Name, v.Value)
	}
}
