OpenTofu - AWS ECR und GitHub Actions

English quick start and file layout

- Tooling: This stack is managed with OpenTofu (tofu). The configuration is compatible with Terraform as well.
- File layout:
  - versions.tf — Terraform/OpenTofu and AWS provider constraints (region, provider version)
  - variables.tf — all input variables for the stack
  - main.tf — resources, data sources and locals only (ECR, IAM/OIDC, Lambda, EventBridge)
  - outputs.tf — outputs printed after apply
- Typical workflow:
  - tofu init — download providers and initialize the working directory
  - tofu plan — preview changes
  - tofu apply — apply changes

Notes for newcomers to OpenTofu

- OpenTofu is a community-driven fork of Terraform. Command names and HCL language are the same; you can use tofu in place of terraform.
- State is stored locally by default in this folder (terraform.tfstate). For teams/CI, consider a remote backend (e.g. S3 + DynamoDB).
- Provider versions are pinned in versions.tf for reproducibility. Update carefully to avoid breaking changes.

Dieses Verzeichnis enthaelt die OpenTofu-Konfiguration, um ein privates Amazon ECR Repository sowie die IAM/OIDC Anbindung fuer GitHub Actions bereitzustellen. Zusaetzlich ist der Release-Workflow so eingerichtet, dass bei jedem Release ein Docker-Image nach Docker Hub (multi-arch) und ein einzelnes ARM64-Image nach ECR gepusht wird.

Voraussetzungen
- AWS Account und Berechtigungen (fuer den Erstaufbau Admin-Rechte empfohlen)
- Region: eu-central-1 (Frankfurt)
- Tools lokal installiert:
  - OpenTofu (tofu)
  - AWS CLI (aws)
  - Docker (inkl. Buildx fuer optionale lokale Tests; QEMU wird nicht benoetigt)

Enthaltene Ressourcen
- ECR Repository: abfallkalender-api (privat)
- Lifecycle Policy: behaelt die letzten 20 Images
- GitHub OIDC Provider (falls nicht bereits vorhanden)
- IAM Rolle github-actions-ecr-push mit minimalen ECR-Push-Rechten
- Lambda-Ausführung: IAM-Rolle, LogGroup, Lambda-Funktion aus Container-Image, Function URL (öffentlich), EventBridge-Warmup

Ausgaben (tofu apply)
- ecr_repository_url - vollstaendige ECR URL
- github_actions_role_arn - ARN der Rolle fuer GitHub Actions
- lambda_function_name - Name der Lambda-Funktion
- lambda_function_url - Oeffentliche URL der Funktion (direkt nutzbar)

Erste Schritte
```
cd infra
# 1) Provider/Plugins laden
tofu init

# 2) Vorschau
tofu plan

# 3) Anwenden
tofu apply
```

Lambda aus Container-Image bereitstellen
Voraussetzung: Es existiert bereits ein Image im ECR mit dem gewuenschten Tag (z. B. 0.0.20). Der
Release-Workflow pusht nur nach Docker Hub, das ECR-Image legst du manuell an - siehe Abschnitt
"Image ins ECR pushen (manuell)" weiter unten. Die Lambda-Funktion zieht genau dieses Image.

Variablen:
- image_tag (erforderlich): Tag des ECR-Images
- lambda_memory_mb (optional, Default 512)
- lambda_timeout_s (optional, Default 15)
- reserved_concurrency (optional, Default null): Reservierte gleichzeitige Ausführungen
  - null (empfohlen): keine Reservierung setzen (vermeidet Quotenfehler)
  - Zahl > 0: setzt eine feste Reservierung. Achtung: AWS verlangt, dass mindestens 10 unreserviert im Konto verbleiben.

Beispiel:
```
cd infra
tofu apply -var "image_tag=0.0.20"

# Ausgabe enthaelt u. a. die URL
# lambda_function_url = https://xxxxxxxxxxxxxxxx.lambda-url.eu-central-1.on.aws/
```

Testaufruf:
```
curl -i $(tofu output -raw lambda_function_url)
```

Warmup: Ein EventBridge Schedule ruft die Funktion alle 5 Minuten auf, um Kaltstarts zu reduzieren.

Hinweis: OIDC Provider bereits vorhanden?
Wenn in deinem AWS Account der GitHub OIDC Provider schon existiert, kann tofu apply mit einem Konflikt fehlschlagen. In dem Fall importiere den bestehenden Provider vor dem Apply:

```
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
# pruefe, ob der Provider existiert
aws iam list-open-id-connect-providers | jq -r '.OpenIDConnectProviderList[].Arn'

# Import in den Tofu-State (ersetze ACCOUNT_ID ggf. manuell)
tofu import aws_iam_openid_connect_provider.github \
  arn:aws:iam::${ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com

# Danach erneut anwenden
tofu apply
```

GitHub Actions - Konfiguration
Releases werden ueber .github/workflows/release.yml ausgeloest (Actions -> Release -> Run workflow,
Auswahl patch/minor/major). Der Workflow legt Tag und GitHub Release an und pusht das Image
multi-arch (linux/amd64, linux/arm64, linux/arm/v7) nach Docker Hub.

Wichtig: Der Release-Workflow fasst AWS nicht an. Es wird kein Image nach ECR gepusht und keine
Lambda aktualisiert. Der AWS-Weg ist bewusst manuell, siehe naechster Abschnitt.

Erforderliche Secrets im GitHub-Repository
- DOCKER_USERNAME - Docker Hub Benutzername
- DOCKER_PASSWORD - Docker Hub Passwort/Token

Hinweis: Die IAM-Rolle github-actions-ecr-push und der GitHub OIDC Provider bleiben in dieser
Konfiguration bestehen, werden aktuell aber von keinem Workflow mehr genutzt. Sie koennen entfernt
werden, falls der ECR-Push dauerhaft manuell bleiben soll.

Hinweis zu Images/Architekturen
Das Dockerfile cross-kompiliert: die Builder-Stage laeuft immer nativ auf der Build-Plattform
(FROM --platform=$BUILDPLATFORM), Go baut ueber GOOS/GOARCH fuer das Ziel. QEMU wird nicht benoetigt.
Es gibt zwei Targets: runner-standard (Docker Hub) und runner-lambda (inkl. AWS Lambda Web Adapter).
Die Lambda-Konfiguration verwendet arm64.

Image ins ECR pushen (manuell)
Der Release-Workflow pusht nur nach Docker Hub. Fuer ein Lambda-Update baust du das ARM64-Image
mit dem Lambda-Adapter lokal und pusht es selbst. Aus dem Repository-Root:

```
# Variablen
REGION=eu-central-1
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
REPO=abfallkalender-api
VERSION=0.0.20
ECR_URI="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/$REPO"

# Login zu ECR
aws ecr get-login-password --region $REGION | docker login \
  --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com"

# Buildx-Builder mit docker-container-Driver (einmalig, fuer Cross-Builds noetig)
docker buildx create --name abfk --driver docker-container --use 2>/dev/null || docker buildx use abfk

# Build + Push (ARM64, Target runner-lambda)
# --provenance/--sbom/oci-mediatypes aus: AWS Lambda akzeptiert keine
# OCI-Index- und Attestation-Manifeste.
docker buildx build \
  --target runner-lambda \
  --platform=linux/arm64 \
  --build-arg VERSION=$VERSION \
  --provenance=false \
  --sbom=false \
  --output=type=registry,oci-mediatypes=false \
  -t "$ECR_URI:$VERSION" \
  . --push
```

Pruefen, dass genau ein arm64-Manifest ohne OCI-Index angekommen ist:
```
aws ecr describe-images --repository-name $REPO --image-ids imageTag=$VERSION
```

Aufraeumen / Entfernen
Achtung: force_delete = true im ECR-Repo loescht das Repository auch dann, wenn noch Images vorhanden sind.

```
cd infra
tofu destroy
```

Troubleshooting
- Fehler beim OIDC Provider: siehe Abschnitt OIDC Provider bereits vorhanden? und importiere den Provider in den State.
- AccessDeniedException beim Push: Der ECR-Push laeuft lokal mit deinen AWS-CLI-Credentials. Pruefe mit `aws sts get-caller-identity`, mit welcher Identitaet du unterwegs bist, und ob sie ECR-Push-Rechte hat.
- RepositoryNotFoundException: tofu apply wurde evtl. noch nicht ausgefuehrt. ECR Repository zuerst anlegen.
- Falsche Region: Stelle sicher, dass ueberall eu-central-1 verwendet wird (Provider, AWS CLI, CI-Env AWS_REGION).

Lambda-spezifisch:
- Image nicht gefunden: Pruefe, ob der angegebene image_tag im ECR vorhanden ist (gleiches Konto/Region).
- 5xx/Timeouts: Erhoehe `lambda_timeout_s` und/oder `lambda_memory_mb`. Logs unter /aws/lambda/abfallkalender-api pruefen.
- CORS: Die Function URL ist mit offenem CORS fuer GET/HEAD/OPTIONS konfiguriert. Bedarfsgerecht anpassen.
- Fehler bei Reserved Concurrency (InvalidParameterValueException: UnreservedConcurrentExecution): Setze keine Reservierung (Default) oder waehle einen kleineren Wert. Du kannst deine Kontolimits pruefen mit:
  ```
  aws lambda get-account-settings --query '{limits:AccountLimit,usage:AccountUsage}'
  ```
  Stelle sicher, dass nach deiner Reservierung mindestens 10 unreserviert verbleiben. Beispiel ohne Reservierung (empfohlen):
  ```
  tofu apply -var "image_tag=0.0.20"
  ```
  Beispiel mit Reservierung (nur wenn Quote passt):
  ```
  tofu apply -var "image_tag=0.0.20" -var "reserved_concurrency=5"
  ```

Naechste Schritte (spaeter)
- Optionale Begrenzung der Kostenrisiken via Reserved Concurrency / API Gateway / WAF
- Hartere Absicherung der Function URL (z. B. IAM/Function URL Auth, CloudFront vorlagern, WAF)