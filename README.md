[![Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](https://i.creativecommons.org/l/by-nc-sa/4.0/88x31.png)]("http://creativecommons.org/licenses/by-nc-sa/4.0/" "Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License")  
This work is licensed under a Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License.

---

# Projet GO Valentin LEMAIRE - Gaspard O'MAHONY

---

## Description

Ce projet permet de faire de la stéganographie sur des images. Il permet de cacher des messages/images dans des images et de les retrouver.

Il implémente une version locale et une version avec un client et un serveur TCP

## Utilisation

### Version locale

Pour utiliser la version locale, il suffit de lancer le programme avec la commande suivante :

```bash
go run Elp.go
```

### Version TCP

Pour utiliser la version TCP, il faut lancer le serveur avec la commande suivante :

```bash
go run tcp/server/devopssec_server.go
```

Puis lancer le client avec la commande suivante :

```bash
go run tcp/client/devopssec_client.go <encode|decode> <key> <image_file> <data_file>
```
---
## Auteurs

* **Valentin LEMAIRE** - *Initial work* - [Valentin Lemaire](https://github.com/28Pollux28)
* **Gaspard O'MAHONY** - *Initial work* - [gaspardomahony](https://github.com/gaspardomahony)

---

## License

This project is licensed under the terms of the CC BY-NC-SA 4.0 license.

