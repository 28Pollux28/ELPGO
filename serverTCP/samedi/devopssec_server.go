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
		panic(err)
	}
}

const (
	IP2   = "localhost" // IP local
	PORT2 = "8000"      // Port utilisé
)

func main() {

	fmt.Println("Lancement du serveur ...")

	// on écoute sur le port 8000
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", IP2, PORT2))
	gestionErreur(err)

	// On accepte les connexions entrantes sur le port 8000
	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}

	// Information sur les clients qui se connectent
	fmt.Println("Un client est connecté depuis", conn.RemoteAddr())

	gestionErreur(err)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() { // goroutine dédiée à l'entrée utilisateur
		defer wg.Done()
		for {
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			gestionErreur(err)

			conn.Write([]byte(text))
		}
	}()

	go func() { // goroutine dédiée à la reception des messages du serveur
		defer wg.Done()
		for {
			message, err := bufio.NewReader(conn).ReadString('\n')
			gestionErreur(err)

			fmt.Print("client : " + message)
		}
	}()

	wg.Wait()

}
