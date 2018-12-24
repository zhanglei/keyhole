// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
)

var simDocs []bson.M

// initialize an array of documents for simulation test.  If a template is available
// read the sample json and replace them with random values.  Otherwise, use the demo
// example.
func (rn Runner) initSimDocs() {
	var err error
	var sdoc bson.M

	if rn.verbose {
		log.Println("initSimDocs")
	}
	rand.Seed(time.Now().Unix())
	total := 512
	if rn.filename == "" {
		for len(simDocs) < total {
			simDocs = append(simDocs, util.GetDemoDoc())
		}
		return
	}

	if sdoc, err = util.GetDocByTemplate(rn.filename, true); err != nil {
		return
	}
	bytes, _ := json.MarshalIndent(sdoc, "", "   ")
	if rn.verbose {
		log.Println(string(bytes))
	}
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)

	for len(simDocs) < total {
		ndoc := make(map[string]interface{})
		util.RandomizeDocument(&ndoc, doc, false)
		delete(ndoc, "_id")
		ndoc["_search"] = strconv.FormatInt(rand.Int63(), 16)
		simDocs = append(simDocs, ndoc)
	}
}

// PopulateData - Insert docs to evaluate performance/bandwidth
// {
//	favorites: {
//		sports: []
//		cities: []
//	}
//	favoriteSports: []
//	favoriteSports1
//	favoriteSports2
//	favoriteSports3
// }
func PopulateData(uri string, sslCAFile string, sslPEMKeyFile string) error {
	var err error
	var client *mongo.Client
	ctx := context.Background()
	if client, err = mdb.NewMongoClient(uri, sslCAFile, sslPEMKeyFile); err != nil {
		panic(err)
	}
	if err = client.Connect(ctx); err != nil {
		panic(err)
	}
	c := client.Database(SimDBName).Collection(CollectionName)
	btime := time.Now()
	for time.Now().Sub(btime) < time.Minute {
		var contentArray []interface{}
		docidx := 0
		for i := 0; i < 100; i++ {
			contentArray = append(contentArray, simDocs[docidx%len(simDocs)])
			docidx++
		}
		if _, err = c.InsertMany(context.Background(), contentArray); err != nil {
			return err
		}
	}

	return nil
}

// Simulate simulates CRUD for load tests
func (rn Runner) Simulate(duration int, transactions []Transaction) {
	var err error
	var client *mongo.Client
	var ctx = context.Background()
	var isTeardown = false
	var totalTPS int

	if client, err = mdb.NewMongoClient(rn.uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
		panic(err)
	}
	if err = client.Connect(ctx); err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	c := client.Database(SimDBName).Collection(CollectionName)

	for run := 0; run < duration; run++ {
		// be a minute transactions
		stage := "setup"
		if run == (duration - 1) {
			stage = "teardown"
			isTeardown = true
			totalTPS = rn.tps
		} else if run > 0 && run < (duration-1) {
			stage = "thrashing"
			totalTPS = rn.tps
		} else {
			totalTPS = rn.tps / 2
		}

		batchCount := 0
		totalCount := 0
		beginTime := time.Now()
		counter := 0
		for time.Now().Sub(beginTime) < time.Minute {
			innerTime := time.Now()
			txCount := 0
			for time.Now().Sub(innerTime) < time.Second && txCount < totalTPS {
				doc := simDocs[batchCount%len(simDocs)]
				batchCount++
				if isTeardown {
					c.DeleteMany(ctx, bson.M{"_search": doc["_search"]})
				} else if len(transactions) > 0 { // --file and --tx
					txCount += execTXByTemplateAndTX(c, util.CloneDoc(doc), transactions)
				} else if len(transactions) == 0 { // --file
					txCount += execTXByTemplate(c, util.CloneDoc(doc))
				} else if rn.filename == "" {
					txCount += execTXForDemo(c, util.CloneDoc(doc))
				}
				// time.Sleep(1 * time.Millisecond)
			} // for time.Now().Sub(innerTime) < time.Second && txCount < totalTPS
			totalCount += txCount
			counter++
			seconds := 1 - time.Now().Sub(innerTime).Seconds()
			if seconds > 0 {
				time.Sleep(time.Duration(seconds) * time.Second)
			}
		} // for time.Now().Sub(beginTime) < time.Minute

		if rn.verbose {
			log.Println("=>", time.Now().Sub(beginTime), time.Now().Sub(beginTime) > time.Minute,
				totalCount, totalCount/counter < totalTPS, counter)
		}
		tenPctOff := float64(totalTPS) * .95
		if rn.verbose || totalCount/counter < int(tenPctOff) {
			log.Printf("%s average TPS was %d, lower than original %d\n", stage, totalCount/counter, totalTPS)
		}

		seconds := 60 - time.Now().Sub(beginTime).Seconds()
		if seconds > 0 {
			time.Sleep(time.Duration(seconds) * time.Second)
		}
		if rn.verbose {
			log.Println("=>", time.Now().Sub(beginTime))
		}
	} //for run := 0; run < duration; run++

	c.Drop(ctx)
}
