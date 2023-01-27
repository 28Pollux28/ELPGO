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
	IP   = "localhost" // IP local
	PORT = "8000"      // Port utilisé
)

func main() {

	var wg sync.WaitGroup

	// Connexion au serveur
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", IP, PORT))
	gestionErreur(err)

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

		//AUTRE OPTION
		// Faire un grand buffer que l'on remplit petit a petit avec un autre buffer temporaire

		// Crée un fichier texte.txt
		file, err := os.Create("texte.txt")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		for {

			// Recevoir le morceau actuel
			message, err := bufio.NewReader(conn).ReadString('\n')
			gestionErreur(err)

			// Écrit le texte "texte" dans le fichier
			_, err = file.WriteString(message)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		fmt.Println("texte.txt créé et texte écrit avec succès!")
	}()

	wg.Wait()

}
