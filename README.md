# Web service billing
## Introduction
Ce service a été fait pour gérer tout ce qui concerne le paiement des commandes d'un client. Ses fonctions sont de :
- gérer le calcul des prix de notre application
- créer une commande paypal (cf. https://developer.paypal.com/docs/api/orders/v2/)
- attendre la validation de paiement d'un client
- capturer le paiement pour valider la commande (cf. https://developer.paypal.com/docs/api/orders/v2/)
- contacter le service web-order permettant de créer le cluster du client lors de la réception de la commande

## Installation
Ce service a été packagé dans une image distroless puis dans une helmchart. Du fait que certaines configurations sont récupérées dans des variables d'environnement, il sera nécessaire de créer un secret Kubernetes contenant ces dernières. La Helmchart ira lire ce secret afin de former l'environnement du client. 

#### Environnement
|VARIABLE|DESCRIPTION|EXEMPLE|
|----|-----------|-----|
|paypal_client_id|xxxxxxxxxxxxxxxxxxxxxxxx|Client ID de notre compte Paypal|
|paypal_client_secret|xxxxxxxxxxxxxxxxxxxxxxxx|Client Secret de notre compte Paypal|
|web_order_service_url|http://localhost:8010/order|URL du service web order permettant la création d'une commande dans notre application|

Exemple de Manifest Kubernetes pour le secret:

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: web-billing-configuration
  namespace: web-billing
# Les secrets sont ici encodés en Base64
data:
  paypal_client_id: eHh4eHh4eHh4
  paypal_client_secret: eHh4eHh4eHh4
  web_order_service_url: aHR0cDovL2xvY2FsaG9zdDo4MDEwL29yZGVy
```

Pour créer ce secret: 

> kubectl apply -f ./fichier-de-configuration-secret.yaml

Pour déployer le service:

> helm upgrade --install web-billing ./web-billing-chart -f ./web-billing-chart.yaml/values.yaml


## Fonctionnement
Ce service peut se voir en 4 étapes simples:
- récupération des prix
- création d'une commande
- approbation de la commande
- capture du paiement du client


#### I - Récupération des prix
Afin de connaître le prix total selon les informations de la commande, un frontend va pouvoir contacter la route **/order/prices** afin de récupérer les prix fixés de notre application.

#### II - Création d'une commande
Lorsqu'un utilisateur valide sa demande de cluster, la route **/order/create** va être contactée. Cette route va :
1. Récupérer un token d'accès à Paypal via nos identifiants
2. Calculer le prix total en fonction des prix fixés dans notre application
3. Créer une commande sur Paypal avec les informations récupérées dans la requête
4. Répondre à la requête HTTP par l'ID de la commande Paypal créée
5. Créer une go routine attendant l'approbation de la commande

#### III - Approbation de la commande
Une fois la commande créée, l'utilisateur va être redirigé vers la page d'authentification pour paiement de Paypal. Lorsque ce dernier a approuvé la commande, la route **/order/approve** sera contactée afin d'envoyer un signal à la go routine précédemment citée, validant la commande.  

#### IV - Capture de la commande
Une fois approuvée, le paiement peut être capturé sur notre Paypal. Cette tâche est effectuée dans la go routine, et non dans une route séparée afin paralléliser les traitements pour différents clients de façon consistante.

## Les routes
Comme évoqué précédemment, il y a 3 routes majeures exposées par ce service. 

### [GET] /order/prices
> Content-Type: application/json 

**HTTP RESPONSE ARGS**
|NOM|DESCRIPTION|
|----|-------------|
|basic|Prix par défaut lors de la commande|
|img_storage_price_unit|Prix par Go de stockage pour les images|
|monitoring_storage_price_unit|Prix par Go de stockage pour le monitoring|
|monitoring_option|Prix d'activation du monitoring|
|alerting_option|Prix d'activation de l'alerting|


### [POST] /order/create
> Content-Type: application/json

**REQUEST BODY**

|NOM|DESCRIPTION|
|------|-------------|
|order_details.cluster_name|(string) Nom du cluster devant suivre la RFC 1123|
|order_details.user_id|(string)ID utilisateur devant suivre la nomenclature xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx|
|order_details.has_monitoring|(bool) Activation du monitoring pour le tenant|
|order_details.has_alerting|(bool)Activation de l'alerting pour le tenant|
|order_details.images_storage|(int) Stockage alloué aux images du tenant (Go)|
|order_details.monitoring_storage|(int) Stockage alloué au monitoring du tenant (Go)|
|currency|(string) Code de la monnaie utilisée pour le paiement (e.g. "EUR")|

**HTTP RESPONSE ARGS**
|NOM|DESCRIPTION|
|----|-------------|
|order_id|ID de la commande créée par Paypal|
|status|Statut de création de commande retourné par Paypal|

### [POST] /order/approve
> Content-Type: application/json

**REQUEST BODY**

|NOM|DESCRIPTION|
|------|-------------|
|order_id|(string) ID de la commande Paypal|

**HTTP RESPONSE ARGS**

|NOM|DESCRIPTION|
|----|-------------|
|id|ID de la commande Paypal|
|status|Statut retourné par Paypal lors de l'approbation de commande|


## TODO
[] Créer une route pour les probes Kubernetes. Cette route doit vérifier dans des go routines séparées : la bonne configuration de l'application, la connexion au service web order. (sleep 30 secondes pour éviter de surcharger l'application)

[] Enlever l'élément "links" lors de la réponse http à order/create

[x] Créer la route pour donner les prix de base

[] Gérer les erreurs lors du calcul de l'order

[] Gérer toutes les réponses HTTP 

[] Vérifier le statut de la commande lors de l'approbation avant de fermer le channel

[] Ajouter des contextes aux requêtes

[] Ajouter les validators sur les champs

