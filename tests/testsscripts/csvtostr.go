package testsscripts

import (
	"decentraland-data-downloader-v4/packages/csv"
	"fmt"
	"log"
	"time"
)

type Person struct {
	Name  string     `json:"name"`
	Age   int        `json:"age"`
	Score float64    `json:"score"`
	Born  *time.Time `json:"born"`
}

func TestCsvToStr() {
	// Using a slice of struct pointers
	var people []*Person
	err := csv.ReadToStruct("testsfiles/people.csv", &people)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range people {
		if p.Born != nil {
			fmt.Printf("%s, age %d, born %s\n", p.Name, p.Age, p.Born.Format(time.RFC3339))
		} else {
			fmt.Printf("%s, age %d, born unknown\n", p.Name, p.Age)
		}
	}
}
