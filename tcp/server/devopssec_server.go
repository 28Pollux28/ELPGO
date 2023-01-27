package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

func gestionErreur(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

const (
	IP2   = "localhost" // IP local
	PORT2 = "8000"      // Port utilisé
)

func main() {

	fmt.Println("Lancement du serveur ...")
	var wg sync.WaitGroup
	go func() { // goroutine dédiée à l'entrée utilisateur
		defer wg.Done()
		for {
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if text == "exit\r\n" {
				fmt.Println("Fermeture du serveur")
				os.Exit(0)
			}
		}
	}()

	// on écoute sur le port 8000
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", IP2, PORT2))
	if err != nil {
		fmt.Println(err)
	}

	// On accepte les connexions entrantes sur le port 8000
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Information sur les clients qui se connectent
		fmt.Println("Un client est connecté depuis", conn.RemoteAddr())
		go func() { // goroutine dédiée à la reception des messages du serveur
			for {
				message, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println("Client déconnecté")
					break
				}
				fmt.Print("client : " + message)
			}
		}()
	}
}
