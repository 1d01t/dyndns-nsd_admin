package main

import (
         _ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"database/sql"
	"time"
	"strings"
	"strconv"
)


// add fqdn if missing
func AddFQDN(domainname string)(string) {
	if !strings.Contains(domainname, Conf.Domain.Name) {
		domainname = domainname + Conf.Domain.Name
	}
	return domainname
}

// check if tables are present
func checkTablePresence(table string) (bool, error){
	var tableExist []byte
	tableExistRow := db.QueryRow("select to_regclass ($1);", table)

	if err := tableExistRow.Scan(&tableExist); err != nil {
		return false, err
	}
	if tableExist != nil {
		return true, nil
	} else {
		return false, nil
	}
}

// create tables in db, if not present
func initializeTables()(bool, bool, bool, error) {
	createdUsers := false
	createdHosts := false
	createdBan := false

	// check if table users exists
	present, err := checkTablePresence("users")
	if err != nil {
		return createdUsers, createdHosts, createdBan, err
	}
	if present == false {
		// create table users
		_, err := db.Exec("CREATE TABLE users (username text primary key, password text)")
		if err != nil {
			return createdUsers, createdHosts, createdBan, err
		}
		createdUsers = true
	}

	// check if table hosts exists
	present, err = checkTablePresence("hosts")
	if err != nil {
		return createdUsers, createdHosts, createdBan, err
	}
	if present == false {
		// create table hosts
		_, err := db.Exec("CREATE TABLE hosts (domainname text primary key, username text, ip text, time timestamp)")
		if err != nil {
			return createdUsers, createdHosts, createdBan, err
		}
		createdHosts = true
	}

	// check if table ban exists
        present, err = checkTablePresence("ban")
	if err != nil {
		return createdUsers, createdHosts, createdBan, err
	}
	if present == false {
                // create table hosts
                _, err := db.Exec("CREATE TABLE ban (ip text primary key, count int, bantime bigint)")
                if err != nil {
			return createdUsers, createdHosts, createdBan, err
		}
		createdBan = true
	}
	return createdUsers, createdHosts, createdBan, nil
}

// add ne entry to database if not exist
func AddToDb(table string, name string, var1 string)(bool, error) {
	var exists string
	var err error
	// check for existence
	switch table {
		case "users":
			err = db.QueryRow("select username from users where username=$1", name).Scan(&exists)
		case "hosts":
			err = db.QueryRow("select domainname from hosts where domainname=$1", name).Scan(&exists)
		default:
			return false, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if len(exists) != 0 {
		// there is already an entry present
		return false, nil
	}
	// entry does not exist. Adding to table
	switch table {
		case "users":
			_, err = db.Exec("insert into users values ($1, $2)", name, var1)
		case "hosts":
			_, err = db.Exec("insert into hosts values ($1, $2, 0, $3)", name, var1, time.Now())
		default:
			return false, nil
	}
	if err != nil {
		// error while writing to db
		return false, err
	}
	return true, nil
}

// add new domain to user
func AddDomain(username string, domainname string)(bool, error) {
	//add fqdn if missing
	domainname = AddFQDN(domainname)

	// save data in users db
	created, err := AddToDb("hosts", domainname, username)
	if err != nil {
		// error while connecting to db
		return false, err
	}
	return created, nil
}

// delete a domain
func DeleteDomain(domainname string)(bool, error) {
	//add fqdn if missing
	domainname = AddFQDN(domainname)

	// check if domainname exists
	var exists string
	err := db.QueryRow("select domainname from hosts where domainname=$1", domainname).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if len(exists) == 0 {
		// domainname not present
		return false, nil
	}
	// remove from hosts table
	_, err = db.Exec("delete from hosts where domainname = $1;", domainname)
	if err != nil {
		// error while connecting to db
		return false, err
	}
	// remove domain from zone file
	err = removeDomainEntryInZone([]string{domainname})
	if err != nil {
		// error while writing to zonefile
		return false, err
	}
	return true, nil
}

// add new user to table users if not exists
func NewUser(username string, password string)(bool, error) {
	// Salt and hash the password
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		// error while generate hash
		return false, err
	}

	// save data in users db
	created, err := AddToDb("users", username, string(hashPwd))
	if err != nil {
		// error while connecting to db
		return false, err
	}
	return created, nil
}

// delete a user from table users
func DeleteUser(username string)(bool, error) {
	// first remove all domains from user in zone file
	delDomains, err := db.Query("select domainname from hosts where username = $1;", username)
	if err != nil {
		// error while reading from db
		return false, err
	}
	// convert domains list from database format to array list
	var delDomainsArray []string
	for delDomains.Next() {
		var hostname string
		if err := delDomains.Scan(&hostname); err != nil {
			return false, err
		}
		delDomainsArray = append(delDomainsArray, hostname)
	}
	if err := delDomains.Err(); err != nil {
		return false, err
	}
	// remove domains from zone file
	err = removeDomainEntryInZone(delDomainsArray)
	if err != nil {
		// error while writing to zonefile
		return false, err
	}

	// remove entries in users db
	_, err = db.Exec("delete from users where username = $1;", username)
	if err != nil {
		// error while writing to db
		return false, err
	}
	// remove entries in hosts table
	_, err = db.Exec("delete from hosts where username = $1;", username)
	if err != nil {
		// error while writing to db
		return false, err
	}
	return true, nil
}

// unban ip
func unbanIP(ip string) (bool, error){
	// check if ip is in ban table
	var exists string
	err := db.QueryRow("select ip from ban where ip = $1", ip).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		// error while connecting to db
		return false, err
	}
	if err == sql.ErrNoRows {
		// no ip found
		return false, nil
	}

	// remove ip in ban table
	_, err = db.Exec("delete from ban where ip = $1;", ip)
	if err != nil {
		// error while connecting to db
		return false, err
	}
	return true, nil
}

// list all users and their domains, IP and timestamp
func listUsers() ([][][]string, error){
	// 3D array; first are all users, second are all domains per user and third are ip and timestamp per domain
	users := [][][]string{}

	// get all users from users table
	allUsers, err := db.Query("select username from users")
	if err != nil {
		// error while reading from db
		return nil, err
	}

	// go thru every single user
	for allUsers.Next() {
		var username string
		if err := allUsers.Scan(&username); err != nil {
			// error while getting usernames
			return nil, err
		}

		// define a 2D array for every single user and add username to first array field
		line := [][]string{}
		line = append (line, []string{username,""})

		// for every user get all domains from hosts table
		allDomains, err := db.Query("select domainname from hosts where username = $1;", username)
		if err != nil {
			// error while reading from db
			return nil, err
		}
		// go thru every single domain of the current user
		for allDomains.Next() {
			//  line for a single domain append with ip and timestamp
			domline := []string{}
			var domain string
			if err := allDomains.Scan(&domain); err != nil {
				// error while getting domainnames
				return nil, err
			}
			// add domain to domainLine
			domline = append (domline, domain)
			// get ip of domainname
			var ip string
			err = db.QueryRow("select ip from hosts where domainname = $1;", domain).Scan(&ip)
			if err != nil {
				// error while reading from db
				return nil, err
			}
			// add ip to domainLine
			domline = append (domline, ip)

			// get timestmp of domain
			var time string
			err = db.QueryRow("select time from hosts where domainname = $1;", domain).Scan(&time)
			if err != nil {
				// error while reading from db
				return nil, err
			}
			// add time to domainLine
			domline = append (domline, strings.Split(strings.Replace(time, "T", " ", -1), ".")[0])
			// add domain, ip and time to domainLine
			line = append (line, domline)
		}
		//add 2D user array to 3D userS array
		users = append(users, line)
	}
	return users, nil
}

// list banned IPs
func listBan() ([][]string, error){
	// twodimensional array for ip, failed tryes and ban time
	ban := [][]string{}

	// get all ips
	allIps, err := db.Query("select ip from ban")
	if err != nil {
		// error while reading from db
		return nil, err
	}

	// go thru every single ip of ban table
	for allIps.Next() {
		var ip string
		if err := allIps.Scan(&ip); err != nil {
			// error while getting ips
			return nil, err
		}

		// line for array with ip in first place, failed tryes 2nd and ban time 3rd
		line := []string{ip}

		// get failed tryes for ip
		var tries string
		err := db.QueryRow("select count from ban where ip = $1;", ip).Scan(&tries)
		if err != nil {
			// error while reading from db
			return nil, err
		}
		// add failed tries to line
		line = append (line, tries)

		// get bantime for ip
		var bantime int64
		err = db.QueryRow("select bantime from ban where ip = $1;", ip).Scan(&bantime)
		if err != nil {
			// error while reading from db
			return nil, err
		}

		// bantime from table minus actual time
		bantime = bantime - time.Now().Unix()

		// add bantime to line
		if bantime <= 0 {
			line = append (line, "is not banned")
		} else {
			line = append (line, strconv.FormatInt((bantime/60), 10)+" minutes")
		}

		// add line to array
		ban = append(ban, line)
	}
	return ban, nil
}

// move domain to another user
func MoveDomain(domainname, username string)(bool, bool, error) {
	//add fqdn if missing
	domainname = AddFQDN(domainname)

	// check if domain exists
	var exists string
	err := db.QueryRow("select domainname from hosts where domainname=$1", domainname).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		// error while connecting to db
		return false, false, err
	}
	if len(exists) == 0 {
		// domain does not exist
		return false, false, nil
	}
	// reset control var
	exists = ""

	// check if user exists
	err = db.QueryRow("select username from users where username=$1", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		// error while connecting to db
		return false, false, err
	}
	if len(exists) == 0 {
		// user does not exist
		return true, false, nil
	}

	// move domain authority to another user
	_, err = db.Exec("update hosts set username=$1 where domainname=$2;", username, domainname)
	if err != nil {
		return false, false, err
	}
	return true, true, nil
}
