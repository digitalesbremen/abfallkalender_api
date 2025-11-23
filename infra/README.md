OpenTofu - AWS ECR und GitHub Actions

Dieses Verzeichnis enthaelt die OpenTofu-Konfiguration, um ein privates Amazon ECR Repository sowie die IAM/OIDC Anbindung fuer GitHub Actions bereitzustellen. Zusaetzlich ist der CI-Workflow so eingerichtet, dass bei neuen Git-Tags ein Docker-Image nach Docker Hub (multi-arch) und ein einzelnes ARM64-Image nach ECR gepusht wird.

Voraussetzungen
- AWS Account und Berechtigungen (fuer den Erstaufbau Admin-Rechte empfohlen)
- Region: eu-central-1 (Frankfurt)
- Tools lokal installiert:
  - OpenTofu (tofu)
  - AWS CLI (aws)
  - Docker (inkl. Buildx/QEMU fuer optionale lokale Tests)

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
Voraussetzung: Es existiert bereits ein Image im ECR mit dem gewuenschten Tag (wird normalerweise durch den CI-Workflow beim Taggen eines Releases gepusht, z. B. v1.2.3). Die Lambda-Funktion zieht genau dieses Image.

Variablen:
- image_tag (erforderlich): Tag des ECR-Images
- lambda_memory_mb (optional, Default 512)
- lambda_timeout_s (optional, Default 15)

Beispiel:
```
cd infra
tofu apply -var "image_tag=v1.2.3"

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
Der Workflow .github/workflows/docker.yml ist so eingerichtet, dass bei jedem Git-Tag:
- nach Docker Hub multi-arch (linux/amd64, linux/arm64, linux/arm) mit latest und dem Versionstag gepusht wird,
- nach ECR ein einzelnes linux/arm64-Image mit dem Versionstag gepusht wird (kein latest).

Erforderliche Secrets/Settings im GitHub-Repository
- DOCKER_USERNAME - Docker Hub Benutzername
- DOCKER_PASSWORD - Docker Hub Passwort/Token
- AWS_ACCOUNT_ID - deine AWS Account ID 

GitHub Actions benoetigt keine AWS Access Keys; der Login erfolgt ueber OIDC und die Rolle github-actions-ecr-push.

Hinweis zu Images/Architekturen
Das Dockerfile baut Images fuer linux/arm64 und linux/amd64; die Lambda-Konfiguration verwendet die Architektur arm64. Stelle sicher, dass fuer den gewaehlten Tag ein arm64-Image im ECR liegt (der CI-Workflow pusht ein einzelnes ARM64-Image ins ECR).

Manuell: Image ins ECR pushen (optional)
Falls du nicht auf die CI warten moechtest, kannst du lokal ein einzelnes ARM64-Image bauen und pushen:

```
# Variablen
REGION=eu-central-1
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
REPO=abfallkalender-api
VERSION=test-local
ECR_URI="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/$REPO"

# Login zu ECR
aws ecr get-login-password --region $REGION | docker login \
  --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com"

# Build (ARM64)
docker buildx build \
  --platform=linux/arm64 \
  --build-arg VERSION=$VERSION \
  -t "$ECR_URI:$VERSION" \
  . --push
```

Aufraeumen / Entfernen
Achtung: force_delete = true im ECR-Repo loescht das Repository auch dann, wenn noch Images vorhanden sind.

```
cd infra
tofu destroy
```

Troubleshooting
- Fehler beim OIDC Provider: siehe Abschnitt OIDC Provider bereits vorhanden? und importiere den Provider in den State.
- AccessDeniedException beim Push aus CI: Pruefe, ob die Rolle github-actions-ecr-push existiert und der Trust auf dein Repo verweist (repo:digitalesbremen/abfallkalender_api:ref:refs/tags/*).
- RepositoryNotFoundException: tofu apply wurde evtl. noch nicht ausgefuehrt. ECR Repository zuerst anlegen.
- Falsche Region: Stelle sicher, dass ueberall eu-central-1 verwendet wird (Provider, AWS CLI, CI-Env AWS_REGION).

Lambda-spezifisch:
- Image nicht gefunden: Pruefe, ob der angegebene image_tag im ECR vorhanden ist (gleiches Konto/Region).
- 5xx/Timeouts: Erhoehe `lambda_timeout_s` und/oder `lambda_memory_mb`. Logs unter /aws/lambda/abfallkalender-api pruefen.
- CORS: Die Function URL ist mit offenem CORS fuer GET/HEAD/OPTIONS konfiguriert. Bedarfsgerecht anpassen.

Naechste Schritte (spaeter)
- Optionale Begrenzung der Kostenrisiken via Reserved Concurrency / API Gateway / WAF
- Hartere Absicherung der Function URL (z. B. IAM/Function URL Auth, CloudFront vorlagern, WAF)