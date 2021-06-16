package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	maxCP  int
}

func ScrapeMoves() {
	doc := prepDoc("https://www.serebii.net/pokemongo/moves.shtml")
	db, err := sql.Open("sqlite3", "file:pokedb.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	tx.Exec("delete from FAST_MOVES")
	tx.Exec("delete from CHARGE_MOVES")
	stmt, err := tx.Prepare("insert into FAST_MOVES(NAME, TYPE, DAMAGE, DPS, ENERGY, DURATION, EPS) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	current_tables := doc.Find("li[title=VCurrent]").Find("table[class=tab]")
	fmoves_table := current_tables.First().Children()
	cmoves_table := current_tables.Last().Children()
	re := regexp.MustCompile(`\d+(\.\d+)?`)
	fmoves_table.Children().Each(func(i int, s *goquery.Selection) {
		if i < 1 {
			return
		}
		cell := s.Children().First()
		name := strings.TrimSpace(cell.Text())
		fmt.Println(name)
		cell = cell.Next()
		url, _ := cell.Children().First().Children().First().Attr("src")
		url_split := strings.Split(url, "/")
		move_type := strings.Split(url_split[len(url_split)-1], ".")[0]
		cell = cell.Next()
		damage, err := strconv.Atoi(strings.TrimSpace(cell.Text()))
		if err != nil {
			log.Fatal(err)
		}
		cell = cell.Next()
		energy, err := strconv.Atoi(strings.TrimSpace(cell.Text()))
		if err != nil {
			log.Fatal(err)
		}
		cell = cell.Next()
		duration, err := strconv.ParseFloat(re.FindString(cell.Text()), 64)
		if err != nil {
			log.Fatal(err)
		}
		mps := 1 / duration
		dps := float64(damage) * mps
		dps, err = strconv.ParseFloat(fmt.Sprintf("%.2f", dps), 64)
		if err != nil {
			log.Fatal(err)
		}
		eps := float64(energy) * mps
		eps, err = strconv.ParseFloat(fmt.Sprintf("%.2f", eps), 64)
		if err != nil {
			log.Fatal(err)
		}
		stmt.Exec(name, move_type, damage, dps, energy, duration, eps)
	})
	stmt, err = tx.Prepare("insert into CHARGE_MOVES(NAME, TYPE, DAMAGE, DPS, ENERGY, DURATION, DPE) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CHARGED")
	cmoves_table.Children().Each(func(i int, s *goquery.Selection) {
		if i < 1 {
			return
		}
		cell := s.Children().First()
		name := strings.TrimSpace(cell.Text())
		fmt.Println(name)
		cell = cell.Next()
		url, _ := cell.Children().First().Children().First().Attr("src")
		url_split := strings.Split(url, "/")
		move_type := strings.Split(url_split[len(url_split)-1], ".")[0]
		cell = cell.Next()
		damage_string := strings.TrimSpace(cell.Text())
		var damage int
		if damage_string == "" {
			damage = 0
		} else {
			damage, err = strconv.Atoi(damage_string)
			if err != nil {
				log.Fatal(err)
			}
		}
		cell = cell.Next().Next()
		duration, err := strconv.ParseFloat(re.FindString(cell.Text()), 64)
		if err != nil {
			log.Fatal(err)
		}
		cell = cell.Next()
		var energy int
		if cell.Children().Length() == 0 {
			energy = 0
		} else {
			energy_url, _ := cell.Children().First().Attr("src")
			if energy_url == "1energy.png" {
				energy = 100
			} else if energy_url == "2energy.png" {
				energy = 50
			} else if energy_url == "3energy.png" {
				energy = 33
			} else if energy_url == "5energy.png" {
				energy = 20
			} else {
				log.Fatal("Unrecognised energy signature:", energy_url)
			}
		}
		mps := 1 / duration
		dps := float64(damage) * mps
		dps, err = strconv.ParseFloat(fmt.Sprintf("%.2f", dps), 64)
		if err != nil {
			log.Fatal(err)
		}
		dpe := int(float64(damage) * (100.0 / float64(energy)))
		stmt.Exec(name, move_type, damage, dps, energy, duration, dpe)
	})
	tx.Commit()
}

func prepDoc(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatal(fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status))
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	/*
		f, err := os.Create("moves.html")
		if err != nil {
			log.Fatal(err)
		}
		html, _ := doc.Html()
		f.Write([]byte(html))
	*/
	return doc
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
	stmt, err := tx.Prepare("insert into POKEMON(DEX, NAME, TYPE1, TYPE2, HP, ATK, DEF, MAXCP) values (?, ?, ?, ?, ?, ?, ?, ?)")
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
		stmt.Exec(dex, info.name, info.types[0], info.types[1], info.hp, info.atk, info.def, info.maxCP)
	}
	tx.Commit()
}

func ScrapeMaxCPOnly() {
	db, err := sql.Open("sqlite3", "file:pokedb.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("select DEX from POKEMON")
	if err != nil {
		log.Fatal(err)
	}
	var dexs []int
	for rows.Next() {
		var val int
		err = rows.Scan(&val)
		if err != nil {
			log.Fatal(err)
		}
		dexs = append(dexs, val)
	}
	stmt, err := tx.Prepare("update POKEMON set MAXCP = ? where DEX == ?")
	if err != nil {
		log.Fatal(err)
	}
	for _, dex := range dexs {
		info, err := ScrapePokeInfo(dex)
		if err != nil {
			log.Fatal(err)
		}
		stmt.Exec(info.maxCP, dex)
	}
	tx.Commit()
}

func ScrapeHasMoves() {
	var nameToFMoveId map[string]int = make(map[string]int)
	var nameToCMoveId map[string]int = make(map[string]int)
	fmt.Println("Opening connection...")
	db, err := sql.Open("sqlite3", "file:pokedb.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Building maps...")
	rows, err := db.Query("select ID, NAME, TYPE from FAST_MOVES")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var name string
		var move_type string
		err = rows.Scan(&id, &name, &move_type)
		if err != nil {
			log.Fatal(err)
		}
		name = strings.ToLower(name)
		nameToFMoveId[name] = id
	}
	rows, err = db.Query("select ID, NAME from CHARGE_MOVES")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		name = strings.ToLower(name)
		nameToCMoveId[name] = id
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Removing old data...")
	tx.Exec("delete from HAS_CHARGE_MOVE")
	tx.Exec("delete from HAS_FAST_MOVE")
	stmtf, err := tx.Prepare("insert into HAS_FAST_MOVE(DEX, MOVE_ID) values (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	stmtc, err := tx.Prepare("insert into HAS_CHARGE_MOVE(DEX, MOVE_ID) values (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	rows, err = db.Query("select DEX from POKEMON")
	if err != nil {
		log.Fatal(err)
	}
	var dexs []int
	for rows.Next() {
		var dex int
		rows.Scan(&dex)
		dexs = append(dexs, dex)
	}
	fmt.Println("Executing statements...")
	for _, dex := range dexs {
		fmt.Printf("#%v\n", dex)
		info, err := ScrapePokeInfo(dex)
		if err != nil {
			log.Fatal(err)
		}
		for _, move := range info.fmoves {
			if move == "" {
				continue
			}
			move_id, ok := nameToFMoveId[move]
			if !ok {
				log.Fatalf("Missing value: %v", move)
			}
			stmtf.Exec(dex, move_id)
		}
		for _, move := range info.cmoves {
			if move == "" {
				continue
			}
			move_id, ok := nameToCMoveId[move]
			if !ok {
				log.Fatalf("Missing value: %v", move)
			}
			stmtc.Exec(dex, move_id)
		}
	}
	fmt.Println("Committing transaction...")
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
	var infoTab, fmovesTab, cmovesTab, statsTab, cpTab *goquery.Selection
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
		case "Flee Rate":
			if cpTab == nil {
				cpTab = s
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
	txt := cpTab.Children().First().Children().Last().Children().Last().Text()
	re, err := regexp.Compile(`\d+`)
	if err != nil {
		log.Fatal(err)
	}
	cp, err := strconv.Atoi(re.FindString(txt))
	if err != nil {
		fmt.Println("CP")
		log.Fatal(err)
	}
	out := PokeInfo{name, types, fmoves, cmoves, stats[0], stats[1], stats[2], cp}
	return out, nil
}
