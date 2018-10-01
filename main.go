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
	Nodes      int      `yaml:"nodes"`
	Metrics    int      `yaml:"metrics"`
	Interval   string   `yaml:"interval"`
	Address    string   `yaml:"address"`
	MainPrefix string   `yaml:"main_prefix"`
	Prefixes   []string `yaml:"prefixes"`
	Values     int      `yaml:"values"`
}

func (conf *spamConfig) String() string {
	prefixes := ""
	for i, pr := range conf.Prefixes {
		if i != 0 {
			prefixes += fmt.Sprintf(", %s", pr)
		} else {
			prefixes += fmt.Sprintf("%s", pr)
		}
	}
	return fmt.Sprintf("Nodes: %d\nMetrics: %d\nInterval: %s\nAddress: %s\nMainPrefix: %s\nPrefixes: %s\n",
		conf.Nodes, conf.Metrics, conf.Interval, conf.Address, conf.MainPrefix, prefixes)
}

func main() {
	flag.Parse()
	config := spamConfig{}

	err := cmd.ReadConfig(*configFileName, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't read settings: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Config:\n=====\n%s", config.String())
	}

	rand.Seed(time.Now().Unix())
	c := graphigo.Client{
		Address: config.Address,
		Timeout: 0,
		Prefix:  config.MainPrefix,
	}

	err = c.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer c.Close()

	ticker := time.NewTicker(to.Duration(config.Interval))
	oldInterval := config.Interval
	configTicker := time.NewTicker(time.Second * 1)
	for {
		select {
		case y := <-ticker.C:
			sendMetrics(c, y, config.Nodes, config.Metrics, config.Values, config.Prefixes)
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

func sendMetrics(c graphigo.Client, y time.Time, nodesCount, metricsCount, values int, prefixes []string) {
	var metrics = make([]graphigo.Metric, 0)
	for _, prefix := range prefixes {
		for i := 0; i < nodesCount; i++ {
			for j := 0; j < metricsCount; j++ {
				name := fmt.Sprintf("Metric_%v.%v", i, j)
				value := rand.Int() % values
				metrics = append(metrics, graphigo.Metric{Name: prefix + "." + name, Value: value})
			}
		}
	}
	c.SendAll(metrics)
	fmt.Printf("%v Metrics: %v\n", y.Format("2006-01-02 15:04:05"), len(metrics))

	for _, v := range metrics {
		fmt.Printf("\t%v.%v - %v\n", c.Prefix, v.Name, v.Value)
	}
}
