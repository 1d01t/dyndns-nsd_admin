package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

//Structs section
type Config struct {
	DB struct {
		Host string `yaml:"host"`
		Port int `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		DBname string `yaml:"dbname"`
	} `yaml:"Database"`
	Files struct {
		LogFile string `yaml:"logfile"`
		ZoneFile string `yaml:"zonefile"`
	} `yaml:"Files"`
	Identifier struct {
		SerialFinder string `yaml:"serialfinder"`
	} `yaml:"Identifier"`
	Singning struct {
		ZoneSigner string `yaml:"zonesigner"`
		KSK string `yaml:"ksk"`
		ZSK string `yaml:"zsk"`
	} `yaml:"Singning"`
	Domain struct {
		Name string `yaml:"name"`
	} `yaml:"Domain"`
}


// Global vars
var Conf Config		// configs from yaml file
var db *sql.DB		// postgres DB
