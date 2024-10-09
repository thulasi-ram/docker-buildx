//go:build !insecure
// +build !insecure

package main

// disallow to change settings that could allow to alter the host running the docker socket
const Insecure = false
