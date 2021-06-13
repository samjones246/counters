package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

type PokeInfo struct {
	name   string
	types  []string
	fmoves []string
	cmoves []string
	hp     int
	atk    int
	def    int
}

func ScrapeAll() {
	db, err := sql.Open("sqlite3", "file:pokedb.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into POKEMON(DEX, NAME, TYPE1, TYPE2, HP, ATK, DEF) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	for dex := 1; dex <= 890; dex++ {
		fmt.Printf("Scraping #%v...\n", dex)
		info, err := ScrapePokeInfo(dex)
		fmt.Println(info)
		if err != nil {
			log.Fatal(err)
		}
		stmt.Exec(dex, info.name, info.types[0], info.types[1], info.hp, info.atk, info.def)
	}
	tx.Commit()
}

func ScrapePokeInfo(dex int) (PokeInfo, error) {
	// Request the HTML page.
	url := fmt.Sprintf("https://www.serebii.net/pokemongo/pokemon/%03d.shtml", dex)
	res, err := http.Get(url)
	if err != nil {
		return PokeInfo{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return PokeInfo{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return PokeInfo{}, err
	}
	s := doc.Find("table[class=dextab]")
	var infoTab, fmovesTab, cmovesTab, statsTab *goquery.Selection
	s.Each(func(i int, s *goquery.Selection) {
		tabTitle := s.Children().First().Children().First().Children().First().Text()
		switch tabTitle {
		case "Picture":
			if infoTab == nil {
				infoTab = s
			}
		case "Fast Attacks":
			if fmovesTab == nil {
				fmovesTab = s
			}
		case "Charge Attacks":
			if cmovesTab == nil {
				cmovesTab = s
			}
		case "Base Stats":
			if statsTab == nil {
				statsTab = s
			}
		}
	})
	row := infoTab.Children().First().Children().First().Next()
	var name string
	var types []string = make([]string, 2)
	var fmoves, cmoves []string
	var stats []int = make([]int, 3)
	row.Children().Each(func(i int, s *goquery.Selection) {
		if i == 1 {
			name = s.Text()
		}
	})
	typesCell := row.Children().Last()
	if typesCell.Children().First().Is("table") {
		typesCell = typesCell.Children().First().Children().First().Children().First().Children().Last()
	}
	typesCell.Children().Each(func(i int, s *goquery.Selection) {
		t, _ := s.Attr("href")
		t = strings.Split(strings.Split(t, "/")[2], ".")[0]
		types[i] = t
	})
	fmovesTab.Children().First().Children().Each(func(i int, s *goquery.Selection) {
		if i > 1 {
			move := s.Children().First().Text()
			move = strings.TrimSpace(move)
			move = strings.ToLower(move)
			fmoves = append(fmoves, move)
		}
	})
	cmovesTab.Children().First().Children().Each(func(i int, s *goquery.Selection) {
		if i > 1 {
			move := strings.TrimSpace(s.Children().First().Text())
			move = strings.ToLower(move)
			cmoves = append(cmoves, move)
		}
	})
	statsTab.Children().First().Children().Last().Children().Each(func(i int, s *goquery.Selection) {
		stats[i], err = strconv.Atoi(s.Text())
		if err != nil {
			log.Fatal(err)
		}
	})
	out := PokeInfo{name, types, fmoves, cmoves, stats[0], stats[1], stats[2]}
	return out, nil
}
