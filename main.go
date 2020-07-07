package main

import (
	"fmt"
	"strconv"
)

func Menu() {
	// clear screen
	print("\033[H\033[2J")

	fmt.Println(" ")
	fmt.Println("     Welcome to Dyndns-Admin area. ")
	fmt.Println("         Please chose action: ")
	fmt.Println(" ")
	fmt.Println(" x________________________________________x")
	fmt.Println(" |                                        |")
        fmt.Println(" |  1  -  Initialize database             |")
	fmt.Println(" |  2     List current users              |")
	fmt.Println(" |  3  -  Add a new user                  |")
	fmt.Println(" |  4     Add a new domainname            |")
	fmt.Println(" |  5  -  Move a domain to another user   |")
	fmt.Println(" |  6     Delete a user and its domains   |")
	fmt.Println(" |  7  -  Delete a domainname             |")
	fmt.Println(" |  8     Unban a banned ip adderss       |")
	fmt.Println(" |  9     List banned IPs                 |")
	fmt.Println(" |  10    Print this menu again           |")
	fmt.Println(" |________________________________________|")
}


func main() {
	// read config file
	initConf()
	// initialize db connection
	initDB()
	// reduce permission rights with pledge
	PledgeAllow()
	// reduce read filesystem rights with unveil
	UnveilAllow()
	// print menu
	var choice string
	Menu()

	for true {
		fmt.Println(" ")
		fmt.Print("> ")
		fmt.Scanln(&choice)
		switch choice {
			case "1": // start initialize DB function
				createdUsers, createdHosts, createdBan, err := initializeTables()
				if err != nil {
					panic(err)
				}
				fmt.Println(" Created table users: "+strconv.FormatBool(createdUsers))
				fmt.Println(" Created table hosts: "+strconv.FormatBool(createdHosts))
				fmt.Println(" Created table ban: "+strconv.FormatBool(createdBan))

			case "2": // start PrintUsers function
				users, err := listUsers()
				if err != nil {
					panic(err)
				}
				for i := range users {
					fmt.Println(" Username: "+users[i][0][0]+"\n")
					for j := range users[i] {
						if j != 0 {
							fmt.Println("     "+users[i][j][0])
							fmt.Println("            "+users[i][j][1])
							fmt.Println("            "+users[i][j][2]+"\n")
						}
					}
					fmt.Println("")
				}

			case "3": // start AddUser function
				var username, password string
				fmt.Print("New Username: ")
				fmt.Scanln(&username)
				fmt.Print("New Password: ")
				fmt.Scanln(&password)
				// check if strings are not empty
				if len(username) == 0 || len(password) == 0 {
					fmt.Println(" You can not leave the input blank. Try again")
					break
				}
				created, err := NewUser(username, password)
				if err != nil {
					panic(err)
				}
				if created == true {
					fmt.Println(" Sucesfully created new user")
				} else {
					fmt.Println(" User already present")
				}

			case "4": // start AddDomain function
				var username, domainname string
				fmt.Print("New Domainname: ")
				fmt.Scanln(&domainname)
				fmt.Print("Username for new domain: ")
				fmt.Scanln(&username)
				// check if strings are not empty
				if len(username) == 0 || len(domainname) == 0 {
					fmt.Println(" You can not leave the input blank. Try again")
					break
				}
				created, err := AddDomain(username, domainname)
				if err != nil {
					panic(err)
				}
				if created == true {
					fmt.Println(" Sucesfully created new domain "+domainname+" for user: "+username)
				} else {
					fmt.Println(" Domain already owned by another user")
				}

			case "5": // start MoveDomain function
				var username, domainname string
				fmt.Print("Domainname to move: ")
				fmt.Scanln(&domainname)
				fmt.Print("Username for domain: ")
				fmt.Scanln(&username)
				// check if strings are not empty
				if len(domainname) == 0 || len(username) == 0 {
					fmt.Println(" You can not leave the input blank. Try again")
					break
				}
				domainFound, userFound, err := MoveDomain(domainname, username)
				if err != nil {
					panic(err)
				}
				if domainFound == true && userFound == true {
					fmt.Println(" Sucesfully moved domain "+domainname+" to user: "+username)
				}
				if domainFound == false {
					fmt.Println(" Domainname not found. Try again")
				}
				if domainFound == true && userFound == false {
					fmt.Println(" Username not found. Try again")
				}

			case "6": // start DelUser function
				var username string
				fmt.Print("Username to delete: ")
				fmt.Scanln(&username)
				// check if string is not empty
				if len(username) == 0 {
					fmt.Println(" You can not leave the input blank. Try again")
					break
				}
				deleted, err := DeleteUser(username)
				if err != nil {
					panic(err)
				}
				if deleted == true {
					fmt.Println(" Sucesfully deleted user: "+username)
				} else {
					fmt.Println(" No such user found to delete: "+username)
				}

			case "7": // start DeleteDomain function
				var domainname string
				fmt.Print("Domainname to delete: ")
				fmt.Scanln(&domainname)
				// check if string is not empty
				if len(domainname) == 0 {
					fmt.Println(" You can not leave the input blank. Try again")
					break
				}
				deleted, err := DeleteDomain(domainname)
				if err != nil {
					panic(err)
				}
				if deleted == true {
					fmt.Println(" Sucesfully deleted domainname: "+domainname)
				} else {
					fmt.Println(" No such domainname found to delete: "+domainname)
				}

			case "8": // start Unban function
				var ip string
				fmt.Print("IP to unban: ")
				fmt.Scanln(&ip)
				// check if string is not empty
				if len(ip) == 0 {
					fmt.Println(" You can not leave the input blank. Try again")
					break
				}
				removed, err := unbanIP(ip)
				if err != nil {
					panic(err)
				}
				if removed == true {
					fmt.Print(" "+ip+" has been unbanned")
				} else {
					fmt.Print(" "+ip+" not found in ban table")
				}

			case "9": // start listBan function
                                ban, err := listBan()
				if err != nil {
					panic(err)
				}
				for i := range ban {
					fmt.Println(" IP:          "+ban[i][0])
					fmt.Println(" Wrong tries: "+ban[i][1])
					fmt.Println(" Ban time:    "+ban[i][2])
					fmt.Println("")
				}

			case "10": // list Menu
				Menu()
			default:
				fmt.Println(" You pressed a wrong button")
		}
		choice = ""
	}
}
