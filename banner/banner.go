package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.2"

func PrintVersion() {
	fmt.Printf("Current haktrailsfree version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
    __            __    __                _  __       ____                
   / /_   ____ _ / /__ / /_ _____ ____ _ (_)/ /_____ / __/_____ ___   ___ 
  / __ \ / __  // //_// __// ___// __  // // // ___// /_ / ___// _ \ / _ \
 / / / // /_/ // ,<  / /_ / /   / /_/ // // /(__  )/ __// /   /  __//  __/
/_/ /_/ \__,_//_/|_| \__//_/    \__,_//_//_//____//_/  /_/    \___/ \___/
`
	fmt.Printf("%s\n%75s\n\n", banner, "Current haktrailsfree version "+version)
}
