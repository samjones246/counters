package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

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
	"fairy":    {"fighting", "dragon", "ghost"},
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

var vuln map[string][]string = map[string][]string{
	"ice":      {"steel", "fire", "fighting", "rock"},
	"dark":     {"bug", "fighting"},
	"normal":   {"fighting"},
	"poison":   {"ground", "psychic"},
	"steel":    {"ground", "fire", "fighting"},
	"fire":     {"ground", "water", "rock"},
	"ground":   {"ice", "water", "grass"},
	"fairy":    {"poison", "steel"},
	"rock":     {"ground", "steel", "water", "grass", "fighting"},
	"water":    {"electric", "grass"},
	"ghost":    {"fairy", "dark", "ghost"},
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

func GetCounters(target string) ([]byte, error) {
	resp, err := http.Get("http://golang.org")
	if err != nil {
		return nil, err
	}
	out := make([]byte, 8)
	resp.Body.Read(out)
	return out, nil
}

func GetTypeCounters(target string) []string {
	return vuln[target]
}

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("usage: counters <type|pokemon>")
	}
	target := strings.ToLower(os.Args[1])
	counters := GetTypeCounters(target)
	fmt.Println(counters)
}