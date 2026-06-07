# Aplikacja Pogodowa — PAwChO Zadanie 2

## Opis rozwiązania

Łańcuch CI/CD (GitHub Actions) budujący obraz kontenera aplikacji pogodowej z zadania 1
i przesyłający go do publicznego repozytorium `ghcr.io/nazln/weather-app-zadanie2`.

Kod źródłowy aplikacji: [`main.go`](main.go)  
Szablon HTML: [`templates/index.html`](templates/index.html)  
Plik workflow: [`.github/workflows/build.yml`](.github/workflows/build.yml)

---

## Plik workflow — build.yml

Pełna treść: [`.github/workflows/build.yml`](.github/workflows/build.yml)

Zastosowane rozwiązania:
- **Wieloplatformowe budowanie** (`linux/amd64` + `linux/arm64`) — za pomocą QEMU i Docker Buildx
- **Cache w rejestrze DockerHub** (`type=registry`, `mode=max`) — przyspiesza kolejne buildy przez przechowywanie warstw na `nazariil/zadanie2-cache:cache`
- **Test CVE przed pushem** — Trivy skanuje obraz lokalnie; pipeline zatrzymuje się przy zagrożeniach CRITICAL lub HIGH, obraz **nie trafia** do ghcr.io
- **Tagowanie SHA + SemVer** — za pomocą `docker/metadata-action@v6`
- **Uprawnienia GITHUB_TOKEN** — `packages: write` umożliwia push do ghcr.io bez dodatkowego tokenu PAT

---

## Konfiguracja środowiska

W repozytorium GitHub ustawiono sekret `DOCKERHUB_TOKEN` (PAT z DockerHub) oraz zmienną `DOCKERHUB_USERNAME`. Na DockerHub utworzono publiczne repozytorium `zadanie2-cache` do przechowywania danych cache.

---

## Schemat tagowania obrazów

Tagowanie realizowane przez `docker/metadata-action@v6`:

| Wyzwalacz | Tag obrazu na ghcr.io |
|-----------|----------------------|
| Push na `main` / ręczne uruchomienie | `sha-<7 znaków SHA>`, np. `sha-a1b2c3d` |
| Push tagu `v1.2.3` | `1.2.3` oraz `1.2` |

```yaml
tags: |
  type=sha,priority=100,prefix=sha-,format=short
  type=semver,priority=200,pattern={{version}}
  type=semver,priority=200,pattern={{major}}.{{minor}}
```

Tag cache na DockerHub: `nazariil/zadanie2-cache:cache` (stały, nadpisywany przy każdym buildzie)

**Uzasadnienie:**  
Schemat SHA + SemVer jest standardową praktyką opisaną w dokumentacji Docker i zgodną ze standardem [Semantic Versioning](https://semver.org/):
- **SHA** zapewnia pełną identyfikowalność — każdy obraz jest powiązany z konkretnym commitem
- **SemVer** umożliwia oznaczanie stabilnych wersji produkcyjnych przy push tagu `vX.Y.Z`
- Wyłączenie `latest` (`flavor: latest=false`) eliminuje ryzyko przypadkowego nadpisania obrazu produkcyjnego

---

## Test CVE — skaner Trivy

Wybrano **Trivy** (Aqua Security) zamiast Docker Scout z następujących powodów:

1. **Open-source** — nie wymaga subskrypcji Docker Pro/Team (Docker Scout w CI jest płatny)
2. **Natywna akcja GitHub** — `aquasecurity/trivy-action` integruje się bezpośrednio z workflow
3. **Proste zatrzymanie pipeline** — `exit-code: '1'` i `severity: CRITICAL,HIGH` blokują push do ghcr.io gdy wykryto zagrożenia
4. **Aktywna baza CVE** — regularnie aktualizowana przez Aqua Security

Skanowanie odbywa się na obrazie zbudowanym lokalnie (**przed** finalnym pushem), co gwarantuje że obraz trafi do ghcr.io **tylko** gdy nie zawiera zagrożeń CRITICAL ani HIGH.

---

## Polecenia

### a. Ręczne uruchomienie pipeline

```bash
gh workflow run build.yml
```

### b. Uruchomienie przez push tagu (SemVer)

```bash
git tag -a "v1.0.0" -m "release v1.0.0"
git push origin tag v1.0.0
```

### c. Obserwacja przebiegu

```bash
gh run watch
```

### d. Sprawdzenie wyników

```bash
# Lista ostatnich runów
gh run list --workflow=build.yml

# Szczegóły ostatniego runu
gh run view
```

Obraz dostępny pod adresem: `ghcr.io/nazln/weather-app-zadanie2`
