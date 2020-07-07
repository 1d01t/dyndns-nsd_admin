package main

import (
	"io/ioutil"
	"strings"
	"errors"
	"regexp"
	"strconv"
	"os/exec"
	"encoding/hex"
	"crypto/rand"
	"crypto/sha256"
	"bytes"
)

// function to remove Domainname entry in zonefile
func removeDomainEntryInZone(delDomains []string) (error) {
	// check if a domain were removed
	DomainnameRemoved := false
        // read zonefile
	zoneFile, err := ioutil.ReadFile(Conf.Files.ZoneFile)
	if err != nil {
		// error while reading zonefile
		return err
	}

	// split file in lines
	lines := strings.Split(string(zoneFile), "\n")

	// i = actual line to work on
	// end = actual max lines in file
	i := 0
	end := cap(lines)

	// check every line in file for domainname and serial number
	for i < end {
		line := lines[i]
		// looking for this comment in zonefile. Before that we have the serial number
		if strings.Contains(line, Conf.Identifier.SerialFinder) {
			// regex to filter only old serial
			re := regexp.MustCompile("[0-9]+")
			// extract serial of zonefile
			serial, err := strconv.Atoi(re.FindAllString(lines[i], -1)[0])
			if err != nil {
				// error by extracting old serial
				return err
			}
			/// increment old serial and replace it
			lines[i] = "	" + strconv.Itoa(serial + 1) + Conf.Identifier.SerialFinder
		} else {
			// check if the line contains one of the "to delete" domain names
			for _, delDomain := range delDomains {
				if strings.Contains(line, delDomain) {
					lines = append(lines[:i], lines[i+1:]...)
					// a line gets removed, so the actual line and the max lines goes -1
					i--
					end--
					DomainnameRemoved = true
				}
			}
		}
		i++
	}
	if DomainnameRemoved == false {
		// domainname was not in zonefile
		return nil
	}
	// merge line to file again
	newZonefile := strings.Join(lines, "\n")
	// write new zonefile
	err = ioutil.WriteFile(Conf.Files.ZoneFile, []byte(newZonefile), 0644)
	if err != nil {
		// error while writing new zonefile
		return err
	}

	// sign new zonefile
	err = SignZone()
	if err != nil {
		// error while executing zone signer
		return err
	}
	return nil
}

// execute cmd commands
func executeCMD(cmd *exec.Cmd)(error) {
	// pipe stdout and stderr to variable
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// execute command
 	if err := cmd.Start(); err != nil {
 		return err
 	}
 	cmd.Wait()

	// handle output of execution
 	if len(stderr.String()) != 0 {
		return errors.New("executeCMD     - STDERR: "+stderr.String())
 	}
	return nil
}

func SignZone()(error) {

	// generate random 32 chars long salt
	randslice := make([]byte, 64)
	if _, err := rand.Read(randslice); err != nil {
		return err
	}
	sha := sha256.Sum256(randslice)
	salt := hex.EncodeToString(sha[:])[0:32]

	// sign zone
	err := executeCMD(exec.Command(Conf.Singning.ZoneSigner, "-n", "-p", "-s", salt, Conf.Files.ZoneFile, Conf.Singning.ZSK, Conf.Singning.KSK))
	if err != nil {
		// error while executing zonesigner
		return err
	}
	err = executeCMD(exec.Command("/usr/sbin/nsd-control", "reload"))
	if err != nil {
		// error while reload nsd zone
		return err
	}
	err = executeCMD(exec.Command("/usr/sbin/nsd-control", "notify"))
	if err != nil {
		// error while notify 2nd dns slave
		return err
	}
	return nil
}
