package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type APIResponse struct {
	Stamps []Stamp `json:"stamps"`
}

type Stamp struct {
	BatchID     string `json:"batchID"`
	Utilization int    `json:"utilization"`
	Usable      bool   `json:"usable"`
	Label       string `json:"label"`
	Depth       int    `json:"depth"`
	//Amount       string `json:"amount"`
	BucketDepth   int  `json:"bucketDepth"`
	BlockNumber   int  `json:"blockNumber"`
	ImmutableFlag bool `json:"immutableFlag"`
	Exists        bool `json:"exists"`
	BatchTTL      int  `json:"batchTTL"`
}

var (
	utilizationMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "swarm_stamp_batch_utilization",
		Help: "Stamp batch utilization.",
	}, []string{"batchID", "label"})
	ttlMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "swarm_stamp_batch_ttl",
		Help: "Stamp batch TTL.",
	}, []string{"batchID", "label"})
	depthMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "swarm_stamp_batch_depth",
		Help: "Stamp batch depth.",
	}, []string{"batchID", "label"})
	// amountMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	//     Name: "swarm_stamp_batch_amount",
	//     Help: "Stamp batch amount.",
	// }, []string{"batchID", "label"})
	capacityMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "swarm_stamp_batch_capacity_bytes",
		Help: "Stamp batch total capacity in bytes.",
	}, []string{"batchID", "label"})
	availabilityMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "swarm_stamp_batch_available_bytes",
		Help: "Stamp batch available capacity in bytes.",
	}, []string{"batchID", "label"})
	usageMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "swarm_stamp_batch_usage_percentage",
		Help: "Stamp batch usage percentage.",
	}, []string{"batchID", "label"})
)

func init() {
	prometheus.MustRegister(utilizationMetric)
	prometheus.MustRegister(ttlMetric)
	prometheus.MustRegister(depthMetric)
	// prometheus.MustRegister(amountMetric)
	prometheus.MustRegister(capacityMetric)
	prometheus.MustRegister(availabilityMetric)
	prometheus.MustRegister(usageMetric)
}

func GetStampMaximumCapacityBytes(depth int) float64 {
	return math.Pow(2, float64(depth)) * 4096
}

func GetStampUsage(utilization, depth, bucketDepth int) float64 {
	return float64(utilization) / math.Pow(2, float64(depth-bucketDepth))

}

func fetchMetrics() {
	url := os.Getenv("BEE_ENDPOINT")
	if url == "" {
		url = "http://localhost:1633/stamps"
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Println("Error decoding JSON data:", err)
		return
	}

	for _, stamp := range apiResponse.Stamps {
		utilizationMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(float64(stamp.Utilization))
		ttlMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(float64(stamp.BatchTTL))
		depthMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(float64(stamp.Depth))
		//amountMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(float64(stamp.Amount))

		capacity := GetStampMaximumCapacityBytes(stamp.Depth)
		capacityMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(capacity)

		stampUsage := GetStampUsage(stamp.Utilization, stamp.Depth, stamp.BucketDepth)
		usageMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(stampUsage)

		available := int(stampUsage * capacity)
		availabilityMetric.With(prometheus.Labels{"batchID": stamp.BatchID, "label": stamp.Label}).Set(float64(available))
	}
}

func main() {
	go func() {
		for {
			fetchMetrics()
			time.Sleep(5 * time.Minute)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":1640", nil))
}
