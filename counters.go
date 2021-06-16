package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

/*
var strong map[string][]string = map[string][]string{
	"normal":   {},
	"fighting": {"normal", "rock", "steel", "ice", "dark"},
	"flying":   {"fighting", "bug", "grass"},
	"poison":   {"grass", "fairy"},
	"ground":   {"poison", "rock", "steel", "fire", "electric"},
	"rock":     {"flying", "bug", "fire", "ice"},
	"bug":      {"grass", "psychic", "dark"},
	"ghost":    {"ghost", "psychic"},
	"steel":    {"rock", "ice", "fairy"},
	"fire":     {"bug", "steel", "grass", "ice"},
	"water":    {"ground", "rock", "fire"},
	"grass":    {"ground", "rock", "water"},
	"electric": {"flying", "water"},
	"psychic":  {"fighting", "poison"},
	"ice":      {"flying", "ground", "grass", "dragon"},
	"dragon":   {"dragon"},
	"fairy":    {"fighting", "dragon", "dark"},
	"dark":     {"ghost", "psychic"},
}

var weak map[string][]string = map[string][]string{
	"normal":   {"rock", "ghost", "steel"},
	"fighting": {"flying", "poison", "psychic", "bug", "ghost", "fairy"},
	"flying":   {"rock", "steel", "electric"},
	"poison":   {"poison", "ground", "rock", "ghost", "steel"},
	"ground":   {"flying", "bug", "grass"},
	"rock":     {"fighting", "ground", "steel"},
	"bug":      {"fighting", "flying", "poison", "ghost", "steel", "fire", "fairy"},
	"ghost":    {"normal", "dark"},
	"steel":    {"steel", "fire", "water", "electric"},
	"fire":     {"rock", "fire", "water", "dragon"},
	"water":    {"water", "grass", "dragon"},
	"grass":    {"flying", "poison", "bug", "steel", "fire", "grass", "dragon"},
	"electric": {"ground", "grass", "electric", "dragon"},
	"psychic":  {"steel", "psychic", "dark"},
	"ice":      {"steel", "fire", "water", "ice"},
	"dragon":   {"steel", "fairy"},
	"fairy":    {"poison", "steel", "fire"},
	"dark":     {"fighting", "dark", "fairy"},
}
*/
var vuln map[string][]string = map[string][]string{
	"ice":      {"steel", "fire", "fighting", "rock"},
	"dark":     {"bug", "fighting", "fairy"},
	"normal":   {"fighting"},
	"poison":   {"ground", "psychic"},
	"steel":    {"ground", "fire", "fighting"},
	"fire":     {"ground", "water", "rock"},
	"ground":   {"ice", "water", "grass"},
	"fairy":    {"poison", "steel"},
	"rock":     {"ground", "steel", "water", "grass", "fighting"},
	"water":    {"electric", "grass"},
	"ghost":    {"dark", "ghost"},
	"bug":      {"flying", "fire", "rock"},
	"grass":    {"ice", "flying", "poison", "fire", "bug"},
	"flying":   {"ice", "electric", "rock"},
	"dragon":   {"ice", "fairy", "dragon"},
	"fighting": {"fairy", "flying", "psychic"},
	"psychic":  {"dark", "bug", "ghost"},
	"electric": {"ground"},
}

var res map[string][]string = map[string][]string{
	"ground":   {"rock", "poison", "electric"},
	"dark":     {"ghost", "psychic", "dark"},
	"water":    {"steel", "fire", "water", "ice"},
	"dragon":   {"fire", "water", "grass", "electric"},
	"rock":     {"normal", "fire", "flying", "poison"},
	"grass":    {"ground", "water", "grass", "electric"},
	"poison":   {"fighting", "poison", "bug", "fairy", "grass"},
	"psychic":  {"fighting", "psychic"},
	"ghost":    {"normal", "fighting", "poison", "bug"},
	"steel":    {"normal", "rock", "steel", "ice", "dragon", "psychic", "flying", "poison", "bug", "fairy", "grass"},
	"flying":   {"ground", "fighting", "bug", "grass"},
	"bug":      {"ground", "fighting", "grass"},
	"ice":      {"ice"},
	"fairy":    {"dragon", "fighting", "dark", "bug"},
	"fighting": {"rock", "dark", "bug"},
	"normal":   {"ghost"},
	"fire":     {"steel", "fire", "ice", "bug", "fairy", "grass"},
	"electric": {"steel", "flying", "electric"},
}

func GetBestAttackers(ptypes []string, db *sqlx.DB) []string {
	q, args, err := sqlx.In(`SELECT P.DEX, P.NAME, P.TYPE1, P.TYPE2, P.MAXCP, FM.NAME AS FAST, MAX(FM.DPS) AS DPS, CM.NAME AS CHARGE, MAX(CM.DPE) AS DPE
	FROM POKEMON P
	INNER JOIN HAS_CHARGE_MOVE HCM
	ON P.DEX == HCM.DEX
	INNER JOIN HAS_FAST_MOVE HFM
	ON P.DEX == HFM.DEX
	LEFT OUTER JOIN CHARGE_MOVES CM
	ON HCM.MOVE_ID == CM.ID
	LEFT OUTER JOIN FAST_MOVES FM
	ON HFM.MOVE_ID == FM.ID
	WHERE (
		TYPE1 IN (?)
	 OR TYPE2 IN (?)
	)
	AND FM.TYPE IN (?)
	AND CM.TYPE IN (?)
	AND (
		FM.TYPE == P.TYPE1
	 OR FM.TYPE == P.TYPE2
	)
	AND (
		CM.TYPE == P.TYPE1
	 OR CM.TYPE == P.TYPE2
	)
	GROUP BY P.DEX
	ORDER BY P.MAXCP DESC
	LIMIT 5`, ptypes, ptypes, ptypes, ptypes)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		log.Fatal(err)
	}
	var out []string
	for rows.Next() {
		var name, t1, t2, fm_name, cm_name string
		var dex, maxcp, cm_dpe int
		var fm_dps float64
		err = rows.Scan(&dex, &name, &t1, &t2, &maxcp, &fm_name, &fm_dps, &cm_name, &cm_dpe)
		if err != nil {
			log.Fatal(err)
		}
		var typestr string
		if t2 != "" {
			typestr = t1 + " and " + t2
		} else {
			typestr = t1
		}
		str := fmt.Sprintf("--%03d %v--\n%v\nMax CP: %v\nFast Move: %v (%.2f DPS)\nCharged Move: %v (%d DPE)",
			dex, name, typestr, maxcp, fm_name, fm_dps, cm_name, cm_dpe)
		out = append(out, str)
	}
	return out
}

func GetTypeCounters(types []string) (two []string, four []string) {
	out := make(map[string]int)
	for _, t := range types {
		for _, tRes := range res[t] {
			out[tRes]--
		}
		for _, tVuln := range vuln[t] {
			out[tVuln]++
		}
	}
	for t, val := range out {
		if val == 1 {
			two = append(two, t)
		}
		if val == 2 {
			four = append(four, t)
		}
	}
	return two, four
}

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	dbpath := os.ExpandEnv("file:${GOPATH}\\bin\\pokedb.sqlite")
	db, err := sqlx.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal(err)
	}
	pname := os.Args[1]
	if pname == "scrape" {
		ScrapeMoves()
		os.Exit(0)
	}
	pname = strings.ToUpper(string(pname[0])) + strings.ToLower(pname[1:])
	rows, err := db.Query("select TYPE1, TYPE2 from 'POKEMON' where NAME LIKE ?", os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	rows.Next()
	var t1, t2 string
	err = rows.Scan(&t1, &t2)
	if err != nil {
		log.Fatalf("Pokemon \"%v\" not found", pname)
	}
	var ts []string
	ts = append(ts, t1)
	if t2 != "" {
		ts = append(ts, t2)
		fmt.Printf("%v is %v and %v type\n", pname, t1, t2)
	} else {
		fmt.Printf("%v is %v type\n", pname, t1)
	}
	two, four := GetTypeCounters(ts)
	fmt.Println("Type Counters:")
	fmt.Println("  1.6x:", two)
	if len(four) > 0 {
		fmt.Println("  2.56x", four)
	}
	var pcounters []string
	if len(four) > 0 {
		pcounters = GetBestAttackers(four, db)
	} else {
		pcounters = GetBestAttackers(two, db)
	}
	for _, counter := range pcounters {
		fmt.Println(counter)
	}
}
