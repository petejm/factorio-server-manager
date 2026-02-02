// use this file only when compiling not windows (all unix systems)
// +build !windows

package factorio

import (
	"log"
	"os"
)

// Stubs for windows-only functions

func (server *Server) Kill() error {
	cmd := server.GetCmd()
	if cmd == nil || cmd.Process == nil {
		log.Printf("No process to kill")
		return nil
	}

	err := cmd.Process.Signal(os.Kill)
	if err != nil {
		if err.Error() == "os: process already finished" {
			server.SetRunning(false)
			return err
		}
		log.Printf("Error sending SIGKILL to Factorio process: %s", err)
		return err
	}
	server.SetRunning(false)
	log.Printf("Sent SIGKILL to Factorio process. Factorio forced to exit.")

	rc := server.GetRcon()
	if rc != nil {
		err = rc.Close()
		if err != nil {
			log.Printf("Error close rcon connection: %s", err)
		}
		server.SetRcon(nil)
	}

	return nil
}

func (server *Server) Stop() error {
	cmd := server.GetCmd()
	if cmd == nil || cmd.Process == nil {
		log.Printf("No process to stop")
		return nil
	}

	err := cmd.Process.Signal(os.Interrupt)
	if err != nil {
		if err.Error() == "os: process already finished" {
			server.SetRunning(false)
			return err
		}
		log.Printf("Error sending SIGINT to Factorio process: %s", err)
		return err
	}
	log.Printf("Sent SIGINT to Factorio process. Factorio shutting down...")

	rc := server.GetRcon()
	if rc != nil {
		err = rc.Close()
		if err != nil {
			log.Printf("Error close rcon connection: %s", err)
		}
		server.SetRcon(nil)
	}

	return nil
}
