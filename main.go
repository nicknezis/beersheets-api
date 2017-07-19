package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"sort"

	"github.com/Luxurioust/excelize"
	"github.com/gorilla/mux"
)

type Ranking struct {
	Name     string  `json:"name"`
	Team     string  `json:"team"`
	Position string  `json:"position"`
	Rank     int     `json:"rank"`
	Bye      int     `json:"bye_week"`
	Tier     float32 `json:"tier"`
	ADP      float32 `json:"average_draft_position"`
}

type BeerSheetRanking struct {
	Name      string `json:"name"`
	Position  string `json:"position"`
	TeamBye   string
	Rank      float32 `json:"rank"`
	AdpVsRank float64
	Past      string
	Value     float32 `json:"value"`
	Scarcity  float32 `json:"scarcity"`
}

type BeerSheetRankings []BeerSheetRanking

var (
	bsPlayers BeerSheetRankings
)

func (slice BeerSheetRankings) Len() int {
	return len(slice)
}

func (slice BeerSheetRankings) Less(i, j int) bool {
	return slice[i].Value > slice[j].Value
}

func (slice BeerSheetRankings) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func main() {
	// xlsx, err := excelize.OpenFile("./data/2016-09-01 10 TM 0 PPR 1QB 2RB 3WR 1TE 1FLX 4 PaTD Snake.xlsx")
	xlsx, err := excelize.OpenFile("./data/2017-07-12 10 TM 0 PPR 1QB 2RB 2WR 1TE 1FLX 4 PaTD Snake.xlsx")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rows := xlsx.GetRows("Sheet1")

	var bsQuarterbacks BeerSheetRankings  //:= []BeerSheetRanking{}  //make([]BeerSheetRanking, 31)
	var bsRunningBacks BeerSheetRankings  //:= []BeerSheetRanking{}  //make([]BeerSheetRanking, 59)
	var bsWideReceivers BeerSheetRankings //:= []BeerSheetRanking{} //make([]BeerSheetRanking, 59)
	var bsTightEnds BeerSheetRankings     //:= []BeerSheetRanking{}

	populateQuarterbacks(rows, &bsQuarterbacks)
	populateRunningBacks(rows, &bsRunningBacks)
	populateWideReceivers(rows, &bsWideReceivers)
	populateTightEnds(rows, &bsTightEnds)

	// fmt.Printf("Players: %v\n", bsQuarterbacks)
	// fmt.Printf("Running Backs: %v\n", bsRunningBacks)
	// fmt.Printf("Wide Receivers: %v\n", bsWideReceivers)
	// fmt.Printf("Tight Ends: %v\n", bsTightEnds)

	bsPlayers = append(bsPlayers, bsQuarterbacks...)
	bsPlayers = append(bsPlayers, bsRunningBacks...)
	bsPlayers = append(bsPlayers, bsWideReceivers...)
	bsPlayers = append(bsPlayers, bsTightEnds...)

	sort.Sort(bsPlayers)
	// fmt.Printf("Players: %v\n", bsPlayers[0:10])

	fmt.Println("Listening on port 8080")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/players", GetPlayers)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func GetPlayers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type playersResponse struct {
		Format      string            `json:"format"`
		TimeUpdated string            `json:"time_updated"`
		Rankings    BeerSheetRankings `json:"rankings"`
	}
	response := &playersResponse{"standard", "now", bsPlayers}
	json.NewEncoder(w).Encode(response)
}

func populateQuarterbacks(rows [][]string, qbs *BeerSheetRankings) {
	for i := 5; i <= 36; i++ {
		row := rows[i]
		// for t := 1; t <= 13; t++ {
		// 	fmt.Println("Row", strconv.Itoa(i), "Col "+strconv.Itoa(t), row[t])
		// }
		var bsPlayer BeerSheetRanking
		bsPlayer.Position = "QB" + strconv.Itoa(i-4)
		bsPlayer.Name = row[2]
		bsPlayer.TeamBye = row[4]
		rank, err := strconv.ParseFloat(row[6], 32)
		if err != nil {
			fmt.Println("Rank error: Row:" + strconv.Itoa(i) + ", Col: 6")
			bsPlayer.Rank = 0
		}
		bsPlayer.Rank = float32(rank)
		adpVsRank, err := strconv.ParseFloat(row[7], 64)
		if err != nil {
			// panic("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 7")
			fmt.Println("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 7")
			bsPlayer.AdpVsRank = 0
		}
		bsPlayer.AdpVsRank = adpVsRank
		bsPlayer.Past = row[8]
		value, err := strconv.ParseFloat(row[10], 32)
		if err != nil {
			panic("Value error: Row:" + strconv.Itoa(i) + ", Col: 10")
		}
		bsPlayer.Value = float32(value)

		scarcity, err := parseFloatPercent(row[13], 32)
		if err != nil {
			fmt.Println("Scarcity error: Row:" + strconv.Itoa(i) + ", Col: 13")
			bsPlayer.Scarcity = 0
		}
		bsPlayer.Scarcity = float32(scarcity)

		*qbs = append(*qbs, bsPlayer)
	}
}

func populateRunningBacks(rows [][]string, rbs *BeerSheetRankings) {
	for i := 5; i <= 64; i++ {
		row := rows[i]
		// for t := 17; t <= 28; t++ {
		// 	fmt.Println("RB: Row", strconv.Itoa(i), "Col "+strconv.Itoa(t), row[t])
		// }
		var bsPlayer BeerSheetRanking
		bsPlayer.Position = "RB" + strconv.Itoa(i-4)
		bsPlayer.Name = row[17]
		bsPlayer.TeamBye = row[19]
		rank, err := strconv.ParseFloat(row[20], 32)
		if err != nil {
			fmt.Println("Rank error: Row:" + strconv.Itoa(i) + ", Col: 20")
			bsPlayer.Rank = 0
		}
		bsPlayer.Rank = float32(rank)
		adpVsRank, err := strconv.ParseFloat(row[21], 64)
		if err != nil {
			// panic("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 7")
			fmt.Println("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 21")
			bsPlayer.AdpVsRank = 0
		}
		bsPlayer.AdpVsRank = adpVsRank
		bsPlayer.Past = row[22]
		value, err := strconv.ParseFloat(row[23], 32)
		if err != nil {
			panic("Value error: Row:" + strconv.Itoa(i) + ", Col: 23")
		}
		bsPlayer.Value = float32(value)

		scarcity, err := parseFloatPercent(row[26], 32)
		if err != nil {
			fmt.Println("Scarcity error: Row:" + strconv.Itoa(i) + ", Col: 26")
			bsPlayer.Scarcity = 0
		}
		bsPlayer.Scarcity = float32(scarcity)

		*rbs = append(*rbs, bsPlayer)
	}
}

func populateWideReceivers(rows [][]string, wrs *BeerSheetRankings) {
	for i := 5; i <= 64; i++ {
		row := rows[i]
		// for t := 30; t <= 42; t++ {
		// 	fmt.Println("WR: Row", strconv.Itoa(i), "Col "+strconv.Itoa(t), row[t])
		// }
		var bsPlayer BeerSheetRanking
		bsPlayer.Position = "WR" + strconv.Itoa(i-4)
		bsPlayer.Name = row[30]
		bsPlayer.TeamBye = row[32]
		rank, err := strconv.ParseFloat(row[34], 32)
		if err != nil {
			fmt.Println("Rank error: Row:" + strconv.Itoa(i) + ", Col: 34")
			bsPlayer.Rank = 0
		}
		bsPlayer.Rank = float32(rank)
		adpVsRank, err := strconv.ParseFloat(row[35], 64)
		if err != nil {
			// panic("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 7")
			fmt.Println("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 35")
			bsPlayer.AdpVsRank = 0
		}
		bsPlayer.AdpVsRank = adpVsRank
		bsPlayer.Past = row[36]
		value, err := strconv.ParseFloat(row[38], 32)
		if err != nil {
			panic("Value error: Row:" + strconv.Itoa(i) + ", Col: 38")
		}
		bsPlayer.Value = float32(value)

		scarcity, err := parseFloatPercent(row[41], 32)
		if err != nil {
			fmt.Println("Scarcity error: Row:" + strconv.Itoa(i) + ", Col: 26")
			bsPlayer.Scarcity = 0
		}
		bsPlayer.Scarcity = float32(scarcity)

		*wrs = append(*wrs, bsPlayer)
	}
}

func populateTightEnds(rows [][]string, tes *BeerSheetRankings) {
	for i := 40; i <= 64; i++ {
		row := rows[i]
		for t := 1; t <= 13; t++ {
			fmt.Println("TE: Row", strconv.Itoa(i), "Col "+strconv.Itoa(t), row[t])
		}
		var bsPlayer BeerSheetRanking
		bsPlayer.Position = "TE" + strconv.Itoa(i-39)
		bsPlayer.Name = row[2]
		bsPlayer.TeamBye = row[4]
		rank, err := strconv.ParseFloat(row[6], 32)
		if err != nil {
			fmt.Println("Rank error: Row:" + strconv.Itoa(i) + ", Col: 6")
			bsPlayer.Rank = 0
		}
		bsPlayer.Rank = float32(rank)
		adpVsRank, err := strconv.ParseFloat(row[7], 64)
		if err != nil {
			// panic("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 7")
			fmt.Println("ADP vs Round error: Row:" + strconv.Itoa(i) + ", Col: 7")
			bsPlayer.AdpVsRank = 0
		}
		bsPlayer.AdpVsRank = adpVsRank
		bsPlayer.Past = row[8]
		value, err := strconv.ParseFloat(row[10], 32)
		if err != nil {
			panic("Value error: Row:" + strconv.Itoa(i) + ", Col: 10")
		}
		bsPlayer.Value = float32(value)

		scarcity, err := parseFloatPercent(row[13], 32)
		if err != nil {
			fmt.Println("Scarcity error: Row:" + strconv.Itoa(i) + ", Col: 13")
			bsPlayer.Scarcity = 0
		}
		bsPlayer.Scarcity = float32(scarcity)

		*tes = append(*tes, bsPlayer)
	}
}

func parseFloatPercent(s string, bitSize int) (f float64, err error) {
	i := strings.Index(s, "%")
	if i < 0 {
		return 0, fmt.Errorf("ParseFloatPercent: percentage sign not found")
	}
	f, err = strconv.ParseFloat(s[:i], bitSize)
	if err != nil {
		return 0, err
	}
	return f / 100, nil
}
