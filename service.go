// +build !windows

package main

//runService dummy for non windows systems
func runService() chan bool {
	return make(chan bool)
}
