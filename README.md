# Web service billing
## Introduction
Ce service a été fait pour gérer tout ce qui concerne le paiement des commandes d'un client. Ses fonctions sont de :
- gérer le calcul des prix de notre application
- créer une commande paypal (cf. https://developer.paypal.com/docs/api/orders/v2/)
- attendre la validation de paiement d'un client
- capturer le paiement pour valider la commande (cf. https://developer.paypal.com/docs/api/orders/v2/)
- contacter le service web-order permettant de créer le cluster du client lors de la réception de la commande


## Les routes
Il y a 3 routes majeures exposées par ce service. 

### [GET] /order/prices
**HTTP BODY**
> Content-Type: application/json 

**HTTP RESPONSE ARGS**
|NOM|DESCRIPTION|
|----|-------------|
|basic|Prix par défaut lors de la commande|
|img_storage_price_unit|Prix par Go de stockage pour les images|
|monitoring_storage_price_unit|Prix par Go de stockage pour le monitoring|
|monitoring_option|Prix d'activation du monitoring|
|alerting_option|Prix d'activation de l'alerting|


### [GET] /order/create
**REQUEST BODY**
> Content-Type: application/json

TODO : Pour la readiness, faire une fonction qui vérifie la bonne connexion avec la bdd + x service nécessaires au bon fonctionnement. 
Ces appels doivent être dans des goroutines qui sleep 30 par exemple histoire de ne pas surcharger le(s) service(s) en cas de multiples appels. 


> TODO:
Enlever le lien d'approval lors du retour de la création de l'order
Faire la méthode de calcul des prix
Faire la route pour donner les prix de base
Trouver un moyen de gérer les erreurs lors du calcul de l'order


> MVP : Faire les méthodes a.validateProbes()

> MVP : Faire les bonnes réponses HTTP (respondwith-xxxxxx) sur les méthodes paypal

> MVP : Vérifier le le statut de l'order (approuvée ou non) avant de fermer le channel (route paypal show order details) 

> ENHANCEMENT: Ajouter des contextes aux requêtes - Très important notamment sur la goroutine qui attend l'approval du client - memory leaks potentiels.

> ENHANCEMENT: Ajouter les validators sur les champs

